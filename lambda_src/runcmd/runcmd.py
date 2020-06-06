import logging
import hashlib
import json
import time
import boto3
import re
import shlex
from os import environ, path
from cgi import escape
from challenge import (
    verify_result,
    value_tests_rand_result,
    value_output_rand_result,
    value_rand_pass,
)
from challenge import bool_to_int_dyn
from docker_cmd import output_from_cmd, DockerValidationError
from dynamo import COMMANDS_TABLE_NAME, SUBMISSIONS_TABLE_NAME
from dynamo import raise_on_rate_limit, DynamoValidationError


LOG = logging.getLogger()
LOG.setLevel(logging.WARN)
DOCKER_HOST = environ["DOCKER_EC2_DNS"]
MAX_OUTPUT_LEN = 10000
SLACK_SNS_ARN = "arn:aws:sns:us-east-1:414252096707:cmdchallenge-slack"


class LambdaValidationError(Exception):
    pass


def send_msg(message=None, channel=None, cmd=None, challenge_slug=None):
    client = boto3.client("sns")
    msg_json = json.dumps(
        dict(message=message, channel=channel, cmd=cmd, challenge_slug=challenge_slug)
    )
    client.publish(TargetArn=SLACK_SNS_ARN, Message=msg_json)


def false_if_empty(input_str):
    if isinstance(input_str, str) and input_str == "":
        return False
    return input_str


def merge_two_dicts(x, y):
    """Given two dicts, merge them into a new dict as a shallow copy."""
    z = x.copy()
    z.update(y)
    return z


def default_resp(err, res=None, cmd=None, challenge_slug=None):
    LOG.debug(f"Responding with: err: {err} res: {res}")
    if err:
        LOG.warn(
            f"Responding with error: err: {err} res: {res} cmd: {cmd} challenge_slug: {challenge_slug}"
        )
        send_msg(str(err), channel="ops", cmd=cmd, challenge_slug=challenge_slug)

    headers = {"Content-Type": "application/json", "Access-Control-Allow-Origin": "*"}

    if err:
        headers.update({"Cache-Control": "no-cache, no-store"})
    else:
        headers.update({"Cache-Control": "max-age=31536000"})

    return {
        "statusCode": "400" if err else "200",
        "body": str(err) if err else json.dumps(res),
        "headers": headers,
    }


def write_item_to_db(item, table_name):
    dynamo_table = boto3.resource("dynamodb").Table(table_name)
    try:
        dynamo_table.put_item(Item=item)
    except Exception as e:
        LOG.exception(f"Error updating the {table_name} table with {item} : {e}")
        raise LambdaValidationError("Database error :(")


def get_cmd_from_db(hashed_submission):
    dynamo_submissions = boto3.resource("dynamodb").Table(COMMANDS_TABLE_NAME)
    try:
        q = dynamo_submissions.get_item(Key={"id": hashed_submission})
    except Exception as e:
        LOG.exception(f"Error querying the database for {hashed_submission}: {e}")
        raise LambdaValidationError("Database error :(")
    if "Item" not in q:
        return None
    cmd = q["Item"]
    json_data = cmd.pop("cmd_json_data", None)
    if json_data:
        cmd = merge_two_dicts(cmd, json.loads(json_data))
    return cmd


def handler(event, context):
    if "queryStringParameters" not in event:
        return default_resp(LambdaValidationError("Missing params."))
    body = event["queryStringParameters"]
    if body is None:
        return default_resp(LambdaValidationError("Missing params."))
    if "challenge_slug" not in body:
        LOG.warn(f"Missing challenge_slug in body: {body}")
        return default_resp(LambdaValidationError("Missing challenge_slug."))
    if "cmd" not in body:
        return default_resp(LambdaValidationError("Missing cmd."))

    challenge_slug = body["challenge_slug"]
    try:
        if not path.exists(f"ch/{challenge_slug}.json"):
            return default_resp(
                LambdaValidationError(f"Invalid challenge: {challenge_slug}")
            )
        with open(f"ch/{challenge_slug}.json") as f:
            challenge = json.load(f)
    except IOError as e:
        LOG.exception(f"Error loading challenge {challenge_slug}: {e}")
        return default_resp(LambdaValidationError("Error loading challenge."))

    version = challenge.get("version", -1)
    cmd = body["cmd"]

    if len(cmd) > 300:
        return default_resp(
            LambdaValidationError("Command is too long."),
            cmd=cmd[:300] + "(truncated)",
            challenge_slug=challenge_slug,
        )

    req_fields = {}

    if "httpMethod" in event:
        # Add additional fields for http requests
        operation = event["httpMethod"]
        if operation != "GET":
            return default_resp(ValueError(f"Unsupported method {operation}"))
        identity = event["requestContext"]["identity"]
        if "headers" in event and "X-Forwarded-For" in event["headers"]:
            req_fields["source_ip"] = event["headers"]["X-Forwarded-For"].split(",")[0]
        else:
            req_fields["source_ip"] = identity["sourceIp"]
        req_fields["user_agent"] = identity["userAgent"]
    else:
        req_fields["user_agent"] = "dummy user agent"
        req_fields["source_ip"] = "1.1.1.1"

    try:
        raise_on_rate_limit(req_fields["source_ip"], "submit_with_cache")
    except DynamoValidationError as e:
        return default_resp(e, cmd=cmd, challenge_slug=challenge_slug)
    # Check to see if request is already in cache
    cmd_shlex = ' '.join(shlex.split(cmd))
    hashed_submission = hashlib.sha256(
        bytes(f"{cmd_shlex}{json.dumps(challenge)}", "utf-8")
    ).hexdigest()
    LOG.debug(f"Checking to see if cmd {cmd} is in the cache")
    try:
        cached_response = get_cmd_from_db(hashed_submission)
    except LambdaValidationError as e:
        return default_resp(e, cmd=cmd, challenge_slug=challenge_slug)
    LOG.debug(f"Cached response is {cached_response}")

    if cached_response is None:
        # not in cache
        # Check to the non-cached rate limit
        LOG.debug(f"Cmd `{cmd}` not in cache")
        LOG.warn(f"New command '{cmd}' with hash '{hashed_submission}'")
        try:
            raise_on_rate_limit(req_fields["source_ip"], "submit_without_cache")
        except DynamoValidationError as e:
            return default_resp(e)

        tls_settings = dict(
            ca_cert="keys/ca.pem",
            verify=True,
            client_cert=("keys/cert.pem", "keys/key.pem"),
        )
        LOG.debug(f"Sending cmd `{cmd}` to docker")
        try:
            result = output_from_cmd(
                cmd,
                challenge,
                docker_version="1.23",
                docker_base_url=f"https://{DOCKER_HOST}:2376",
                tls_settings=tls_settings,
            )
            LOG.warn(
                f"Got result {result} for command '{cmd}' with hash '{hashed_submission}'"
            )
            return_code = result["CmdExitCode"]
            output = result["CmdOut"].rstrip()
            if return_code != 0:
                output = re.sub("^bash: ", "", output)
            test_errors = result.get("TestsOut", None)
            if test_errors is not None:
                # Need to remove all empty strings before storing
                test_errors = list(filter(None, test_errors.strip().split("\n")))
        except DockerValidationError as e:
            return default_resp(e, cmd=cmd, challenge_slug=challenge_slug)
        correct = verify_result(result)
        rand_output_pass = value_output_rand_result(result)
        rand_tests_pass = value_tests_rand_result(result)
        rand_error = not value_rand_pass(result)
        if correct:
            send_msg(channel="commands", cmd=cmd, challenge_slug=challenge_slug)
    else:
        LOG.debug(f"Found cmd `{cmd}` in cache")
        output = str(cached_response["output"])
        return_code = int(cached_response["return_code"])
        test_errors = cached_response["test_errors"]
        correct = cached_response["correct"]
        rand_output_pass = cached_response.get("rand_output_pass", 0)
        rand_tests_pass = cached_response.get("rand_tests_pass", 0)
        rand_error = cached_response.get("rand_error", False)

    create_time = int(time.time() * 100)
    cmd_len = len(cmd)
    correct_length = int(str(bool_to_int_dyn(correct)) + str(cmd_len).zfill(10))
    if len(output) > MAX_OUTPUT_LEN:
        output = output[:MAX_OUTPUT_LEN] + " ** truncated ** "

    item = merge_two_dicts(
        req_fields,
        dict(
            create_time=create_time,
            challenge_slug=challenge_slug,
            correct=correct,
            rand_output_pass=rand_output_pass,
            rand_tests_pass=rand_tests_pass,
            rand_error=rand_error,
            correct_length=correct_length,
            return_code=return_code,
            test_errors=test_errors,
            cmd=false_if_empty(cmd),
            output=false_if_empty(output),
            cmd_length=cmd_len,
            version=version,
            cmd_json_data=json.dumps(dict(cmd=cmd, output=output)),
        ),
    )

    # Add to the submissions table
    try:
        LOG.debug(
            f"Writing `{cmd}` to the submissions table. cached_response: {cached_response}"
        )
        write_item_to_db(item, SUBMISSIONS_TABLE_NAME)
        if cached_response is None:
            # Add to the commands table if it is a new command
            item["id"] = hashed_submission
            LOG.debug(f"Writing `{cmd}` to the commands table")
            write_item_to_db(item, COMMANDS_TABLE_NAME)
    except LambdaValidationError as e:
        return default_resp(e, challenge_slug=challenge_slug, cmd=cmd)

    # Return a successful response
    return default_resp(
        None,
        dict(
            output=escape(output),
            rand_error=rand_error,
            challenge_slug=challenge_slug,
            return_code=return_code,
            correct=correct,
            test_errors=test_errors,
        ),
    )
