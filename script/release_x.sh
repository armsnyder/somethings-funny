#!/usr/bin/env bash
set -e
[[ -e .env ]] && source .env
[[ -z ${AWS_DEFAULT_REGION} ]] && echo AWS_DEFAULT_REGION required && exit 1
[[ -z ${AWS_ACCOUNT_ID} ]] && echo AWS_ACCOUNT_ID required && exit 1
NAME=$1
ECR_REPO=somethings-funny/${NAME}
$(aws ecr get-login --no-include-email)
IMAGE=${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_DEFAULT_REGION}.amazonaws.com/${ECR_REPO}:latest
eval $(docker-machine env -u)
docker build -t ${IMAGE} . -f Dockerfile-${NAME}
docker push ${IMAGE}
IMAGES_TO_DELETE=$(aws ecr list-images --repository-name ${ECR_REPO} \
  --filter "tagStatus=UNTAGGED" --query 'imageIds[*]' --output json )
aws ecr batch-delete-image --repository-name ${ECR_REPO} --image-ids "$IMAGES_TO_DELETE" || true
