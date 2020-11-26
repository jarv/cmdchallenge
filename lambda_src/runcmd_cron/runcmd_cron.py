import logging
import time
import re
import boto3
from boto3.dynamodb.conditions import Key
from io import StringIO
import json
from os import environ
from os.path import join, dirname, realpath

LOG = logging.getLogger()
LOG.setLevel(logging.WARN)
KEY_PREFIX = "s/solutions"
COMMANDS_TABLE_NAME = environ.get(
    "COMMANDS_TABLE_NAME", "prod-cmdchallenge-db-commands"
)
BUCKET_NAME = environ.get("BUCKET_NAME", "testing.cmdchallenge.com")
SHARD_INDEX = int(environ.get("SHARD_INDEX", "0"))
NUM_SHARDS = int(environ.get("SHARD", "10"))
dir_path = dirname(realpath(__file__))
static_path = join(dir_path, "../../static")


def challenges_slice(challenges):
    """
    Given a list, returns a subset based the shard index
    and the number of shards
    """
    last_slug_index = len(challenges) - 1
    start_index = int((last_slug_index * SHARD_INDEX) / NUM_SHARDS)
    if (SHARD_INDEX + 1) == NUM_SHARDS:
        end_index = len(challenges)
    else:
        end_index = int((last_slug_index * (SHARD_INDEX + 1)) / NUM_SHARDS)
    return challenges[start_index:end_index]


def handler(event, context):
    cmds = set()
    challenges = json.loads(open(join(dir_path, "ch/challenges.json")).read())

    if environ.get("LOCAL"):
        b = boto3.session.Session(profile_name="cmdchallenge", region_name="us-east-1")
        s3 = b.client("s3")
        table = b.resource("dynamodb").Table(COMMANDS_TABLE_NAME)
        challenges = challenges_slice(challenges)
    else:
        s3 = boto3.client("s3")
        table = boto3.resource("dynamodb").Table(COMMANDS_TABLE_NAME)
        challenges = challenges_slice(challenges)

    for challenge in challenges:
        slug_name = challenge["slug"]
        resp = table.query(
            IndexName="challenge_slug-correct_length-index",
            KeyConditionExpression=Key("challenge_slug").eq(slug_name)
            & Key("correct_length").lt(20000000000),
            ScanIndexForward=True,
        )
        data = [
            item for item in resp["Items"] if item["version"] == challenge["version"]
        ]
        while "LastEvaluatedKey" in resp:
            resp = table.query(
                ExclusiveStartKey=resp["LastEvaluatedKey"],
                IndexName="challenge_slug-correct_length-index",
                KeyConditionExpression=Key("challenge_slug").eq(slug_name)
                & Key("correct_length").lt(20000000000),
                ScanIndexForward=True,
            )
            data.extend(
                [
                    item
                    for item in resp["Items"]
                    if item["version"] == challenge["version"]
                ]
            )

        cmds = sorted(
            list(set(re.sub(r"\s{2,}", " ", i["cmd"].strip()) for i in data)),
            key=lambda x: len(x),
        )
        LOG.warning(f"Found {len(cmds)} results for slug: {slug_name}")
        results = dict(cmds=cmds, ts=time.time())

        fresults = StringIO(json.dumps(results))
        resp = s3.put_object(
            Bucket=BUCKET_NAME,
            Key=f"{KEY_PREFIX}/{slug_name}.json",
            Body=fresults.read(),
            ACL="public-read",
            CacheControl="no-cache, no-store, must-revalidate",
            ContentType="application/json",
        )
        if resp["ResponseMetadata"]["HTTPStatusCode"] != 200:
            LOG.error(f"Unable to write to s3 bucket {BUCKET_NAME}: {results} {resp}")
            raise Exception(f"Unable to write to S3: {resp}")

        if environ.get("LOCAL"):
            with open(join(static_path, f"{KEY_PREFIX}/{slug_name}.json"), "w") as f:
                f.write(json.dumps(results))


if environ.get("LOCAL"):
    handler(0, 0)
