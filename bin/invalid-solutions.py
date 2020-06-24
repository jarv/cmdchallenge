#!/usr/bin/env python

import logging
import time
import re
import boto3
from boto3.dynamodb.conditions import Key
from io import StringIO
import json
from os import environ
from os.path import join, dirname, realpath
from IPython import embed

LOG = logging.getLogger()
LOG.setLevel(logging.WARN)
COMMANDS_TABLE_NAME = environ.get("COMMANDS_TABLE_NAME", "prod-cmdchallenge-db-commands")
dir_path = dirname(realpath(__file__))

b = boto3.session.Session(profile_name='cmdchallenge', region_name='us-east-1')
table = b.resource('dynamodb').Table(COMMANDS_TABLE_NAME)
slugs = json.loads(open(join(dir_path, "../lambda_src/runcmd_cron/ch/all-challenges.json")).read())

data = {}

for slug_name in ["count_string_in_line"]:
    resp = table.query(
        IndexName="challenge_slug-correct_length-index",
        KeyConditionExpression=Key("challenge_slug").eq(slug_name) & Key("correct_length").gt(20000000000),
        ScanIndexForward=True,
    )
    data = resp["Items"]
    while "LastEvaluatedKey" in resp:
        resp = table.query(
            ExclusiveStartKey=resp["LastEvaluatedKey"],
            IndexName="challenge_slug-correct_length-index",
            KeyConditionExpression=Key("challenge_slug").eq(slug_name) & Key("correct_length").gt(20000000000),
            ScanIndexForward=True,
        )
        data.extend(resp["Items"])
embed()
