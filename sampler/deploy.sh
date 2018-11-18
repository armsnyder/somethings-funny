#!/usr/bin/env bash
set -e

cd "$(dirname "${0}")"
[[ -e "../env.sh" ]] && source ../env.sh
IMAGE="${DOCKER_HUB_USERNAME}/somethings-funny:sampler"
CONTAINER="sampler"
docker build -t "${IMAGE}" .
docker login -u "${DOCKER_HUB_USERNAME}" -p "${DOCKER_HUB_PASSWORD}"
docker push "${IMAGE}"
ssh -o StrictHostKeychecking=no -i "${PI_KEY_PATH}" "${PI_USERNAME}"@"${PI_ADDRESS}" -p "${PI_PORT}" \
  "eval docker pull $IMAGE && \
  docker rm -f $CONTAINER 2>/dev/null || true && mkdir -p sampler/out &&
  docker run -d --restart always --name $CONTAINER --device /dev/snd -v /home/$PI_USERNAME/sampler/out:/out \
  -e AWS_REGION=$AWS_REGION -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e BUCKET_NAME=$BUCKET_NAME $IMAGE"
