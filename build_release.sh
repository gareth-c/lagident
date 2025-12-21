#!/bin/bash

set -e
set -u

DOCKER_USER="nook24"
DOCKER_IMAGE="${DOCKER_USER}/lagident"
VERSION=$(cat VERSION)

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker buildx create --driver docker-container --use
docker buildx inspect --bootstrap

# Keep images only local
#docker buildx build --load --platform linux/amd64,linux/arm64 --tag ${DOCKER_IMAGE}:${VERSION} --tag ${DOCKER_IMAGE}:latest  .

# Push images to registry
docker buildx build --push --platform linux/amd64,linux/arm64 --tag ${DOCKER_IMAGE}:${VERSION} --tag ${DOCKER_IMAGE}:latest  .

# Cleanup (optional)
#docker image prune --filter label=stage=build-lagident-intermediate -f
