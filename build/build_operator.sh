#!/bin/bash

NOW=$(date +%Y%m%d%H%M)
IMG_NAME=ws-demo-operator:${NOW}

ORIGIN_DIR=$(pwd)

# build image
cp ./release/operator ./dockerfile/operator/
cd ./dockerfile/operator/
docker build -t ${IMG_NAME} .
rm ./operator
cd ${ORIGIN_DIR}

# push image
docker login ${Registry} -u ${RegistryAK} -p ${RegistrySK}
docker tag ${IMG_NAME} ${RegistryPrefix}/${IMG_NAME}
docker push ${RegistryPrefix}/${IMG_NAME}
