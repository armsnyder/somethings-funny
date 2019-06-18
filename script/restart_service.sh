#!/usr/bin/env bash
set -e
[[ -e .env ]] && source .env
[[ -z ${AWS_DEFAULT_REGION} ]] && echo AWS_DEFAULT_REGION required && exit 1
[[ -z ${AWS_ACCOUNT_ID} ]] && echo AWS_ACCOUNT_ID required && exit 1
TASK=$(aws ecs list-tasks --service somethings-funny --query taskArns --output text)
aws ecs stop-task --task ${TASK} --reason "Local dev restart service from script" \
  --query task.taskArn --output text
