#!/usr/bin/env bash
set -e

cd "$(dirname "${0}")"
IMAGE="${DOCKER_HUB_USERNAME}/somethings-funny:sampler"
CONTAINER="sampler"
docker build -t "${IMAGE}" .
docker login -u "${DOCKER_HUB_USERNAME}" -p "${DOCKER_HUB_PASSWORD}"
docker push "${IMAGE}"
ssh -o StrictHostKeychecking=no -i "${PI_KEY_PATH}" "${PI_USERNAME}"@"${PI_ADDRESS}" -p "${PI_PORT}" \
  "eval docker pull $IMAGE && \
  docker rm -f $CONTAINER 2>/dev/null || true && mkdir -p sampler/out &&
  docker run -d --restart always --name $CONTAINER --device /dev/snd -v /home/$PI_USERNAME/sampler/out:/out $IMAGE"
