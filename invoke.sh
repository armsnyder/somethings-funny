#!/usr/bin/env bash
set -e
cd "$(dirname "${0}")"
[[ -e "env.sh" ]] && source env.sh
aws sqs send-message --queue-url $SQS_URL --message-body foo --region $AWS_REGION
