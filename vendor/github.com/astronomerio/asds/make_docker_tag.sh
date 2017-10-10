#!/usr/bin/env bash

if [ $# -eq 0 ]; then
    echo "Please supply docker tag"
    exit
fi

TAG=$1

# Tag should start with v!
if [[ "${TAG:0:1}" != "v" ]]; then
    TAG="v$TAG"
fi

make image

docker tag astronomerio/asds:latest astronomerio/asds:$TAG