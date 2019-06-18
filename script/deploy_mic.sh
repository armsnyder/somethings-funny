#!/usr/bin/env bash
set -e
[[ -e .env ]] && source .env
[[ -z ${PI_USERNAME} ]] && echo PI_USERNAME is required && exit 1
IMAGE="somethings-funny:mic"
CONTAINER="mic"
eval $(docker-machine env pi)
docker build -t ${IMAGE} -f Dockerfile-mic .
if [[ -n $(docker ps -q -f name=${CONTAINER}) ]]; then
  echo Stopping previous container
  docker stop ${CONTAINER}
  echo Renaming previous container
  docker rename ${CONTAINER} ${CONTAINER}-$(date +%s)
fi
echo Starting new container
docker run -d --restart unless-stopped --name ${CONTAINER} --device /dev/snd -v /home/${PI_USERNAME}/mic:/mic ${IMAGE}
