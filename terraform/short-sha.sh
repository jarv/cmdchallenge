#!/usr/bin/env sh
jq -n --arg short_sha "$(git rev-parse --short HEAD)" '{"short_sha": $short_sha}'
