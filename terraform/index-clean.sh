#!/usr/bin/env sh
git diff --exit-code 2>&1 >/dev/null && clean="yes" || clean="no"
echo "{\"index_clean\": \"$clean\"}"
