#!/bin/bash

NOW=$(date +%Y%m%d%H%M)
IMG_NAME=simple-ws:${NOW}

ORIGIN_DIR=$(pwd)

# build image
cd ./dockerfile/webserver/
docker build -t ${IMG_NAME} .
rm ./simple_server
cd ${ORIGIN_DIR}

# push image
docker login ${Registry} -u ${RegistryAK} -p ${RegistrySK}
docker tag ${IMG_NAME} ${RegistryPrefix}/${IMG_NAME}
docker push ${RegistryPrefix}/${IMG_NAME}
