#!/usr/bin/env bash
set -e
[[ -e .env ]] && source .env
[[ -z ${PI_USERNAME} ]] && echo PI_USERNAME is required && exit 1
[[ -z ${SERVICE_DOMAIN} ]] && echo SERVICE_DOMAIN is required && exit 1
[[ -z ${SAMPLER_TOKEN} ]] && echo SAMPLER_TOKEN is required && exit 1
[[ -z ${SAMPLER_THRESHOLD} ]] && echo SAMPLER_THRESHOLD is required && exit 1
IMAGE="somethings-funny:sampler"
CONTAINER="sampler"
eval $(docker-machine env pi)
docker build -t ${IMAGE} -f Dockerfile-sampler .
if [[ -n $(docker ps -q -f name=${CONTAINER}) ]]; then
  echo Stopping previous container
  docker stop ${CONTAINER}
  echo Renaming previous container
  docker rename ${CONTAINER} ${CONTAINER}-$(date +%s)
fi
echo Starting new container
docker run -d --restart unless-stopped --name ${CONTAINER} \
  -e UPLOAD_URL=https://${SERVICE_DOMAIN}/upload -e TOKEN=${SAMPLER_TOKEN} \
  -e THRESHOLD=${SAMPLER_THRESHOLD} -v /home/${PI_USERNAME}/mic:/mic ${IMAGE}
