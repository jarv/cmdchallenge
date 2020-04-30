import boto3
import time
from boto3.dynamodb.conditions import Key
import logging
from os import environ

LOG = logging.getLogger()
LOG.setLevel(logging.WARN)
COMMANDS_TABLE_NAME = environ.get("COMMANDS_TABLE_NAME", "commands")
SUBMISSIONS_TABLE_NAME = environ.get("SUBMISSIONS_TABLE_NAME", "submissions")

RATE_LIMITS = {
    "submit_with_cache": dict(num=100, time_span=60),  # 15 * 60 * 60
    "submit_without_cache": dict(num=10, time_span=10),
}


class DynamoValidationError(Exception):
    pass


class commands:
    cmd = "cmd"
    test_errors = "test_errors"
    return_code = "return_code"
    output = "output"
    correct = "correct"


def raise_on_rate_limit(ip, limit_type="submit_with_cache"):
    LOG.debug("Checking request for rate limiting: {}".format(limit_type))
    dynamo_updates = boto3.resource("dynamodb").Table(SUBMISSIONS_TABLE_NAME)
    time_span = int(time.time() * 100) - (RATE_LIMITS[limit_type]["time_span"] * 100)
    c = dynamo_updates.query(
        KeyConditionExpression=Key("source_ip").eq(ip)
        & Key("create_time").gt(time_span)
    )["Count"]

    if c > RATE_LIMITS[limit_type]["num"]:
        LOG.warn("Ip: {} is rate limited {}".format(ip, limit_type))
        raise DynamoValidationError("You are doing too many submissions, slow down")
