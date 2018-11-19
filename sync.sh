#!/usr/bin/env bash
set -e
cd "$(dirname "${0}")"
[[ -e "env.sh" ]] && source env.sh
aws s3 sync s3://$BUCKET_NAME downloads --region $AWS_REGION
aws s3 rm s3://$BUCKET_NAME --recursive --region $AWS_REGION
