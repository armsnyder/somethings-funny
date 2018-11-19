#!/usr/bin/env bash
set -e

cd "$(dirname "${0}")"
[[ -e "../env.sh" ]] && source ../env.sh
IMAGE="somethings-funny:sampler"
CONTAINER="sampler"
ssh -o StrictHostKeychecking=no -i "${PI_KEY_PATH}" "${PI_USERNAME}"@"${PI_ADDRESS}" -p "${PI_PORT}" "eval rm -rf /tmp/sampler"
scp -o StrictHostKeychecking=no -i "${PI_KEY_PATH}" -P "${PI_PORT}" -r "$(pwd)/" "${PI_USERNAME}@${PI_ADDRESS}:/tmp/"
ssh -o StrictHostKeychecking=no -i "${PI_KEY_PATH}" "${PI_USERNAME}"@"${PI_ADDRESS}" -p "${PI_PORT}" \
  "eval cd /tmp/sampler && \
  docker build -t ${IMAGE} . && \
  if [ -n $(docker ps -q -f name=${CONTAINER}) ]; then \
  echo Stopping previous container && \
  docker stop $CONTAINER && \
  echo Renaming previous container && \
  docker rename $CONTAINER $CONTAINER-$(date +%s); fi && \
  echo Starting new container && \
  docker run -d --restart unless-stopped --name $CONTAINER --device /dev/snd -v /home/$PI_USERNAME/mic:/mic \
  -e AWS_REGION=$AWS_REGION -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e SQS_URL=$SQS_URL \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e BUCKET_NAME=$BUCKET_NAME $IMAGE"
