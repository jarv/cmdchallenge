#!/usr/bin/env python

import logging
import boto3
from boto3.dynamodb.conditions import Key
import json
from os import environ
from os.path import join, dirname, realpath
from collections import defaultdict

LOG = logging.getLogger()
LOG.setLevel(logging.WARN)
COMMANDS_TABLE_NAME = environ.get(
    "COMMANDS_TABLE_NAME", "prod-cmdchallenge-db-commands"
)
dir_path = dirname(realpath(__file__))

b = boto3.session.Session(profile_name="cmdchallenge", region_name="us-east-1")
table = b.resource("dynamodb").Table(COMMANDS_TABLE_NAME)
challenges = json.loads(
    open(join(dir_path, "../lambda_src/runcmd_cron/ch/challenges.json")).read()
)

data = []

for challenge in challenges:
    slug_name = challenge["slug"]
    resp = table.query(
        IndexName="challenge_slug-correct_length-index",
        KeyConditionExpression=Key("challenge_slug").eq(slug_name)
        & Key("correct_length").lt(20000000000),
        ScanIndexForward=True,
    )
    data.extend(
        [item for item in resp["Items"] if item["version"] == challenge["version"]]
    )
    while "LastEvaluatedKey" in resp:
        resp = table.query(
            ExclusiveStartKey=resp["LastEvaluatedKey"],
            IndexName="challenge_slug-correct_length-index",
            KeyConditionExpression=Key("challenge_slug").eq(slug_name)
            & Key("correct_length").lt(20000000000),
            ScanIndexForward=True,
        )
        data.extend(
            [item for item in resp["Items"] if item["version"] == challenge["version"]]
        )

d = defaultdict(list)

for i in data:
    d[i["challenge_slug"]].append(i["cmd"])

for k, v in d.items():
    print(f"{k}:")
    for c in v:
        print(f"\t{c}")
