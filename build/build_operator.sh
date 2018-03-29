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
docker login ${QiniuRegistry} -u ${QiniuRegistryAK} -p ${QiniuRegistrySK}
docker tag ${IMG_NAME} ${QiniuRegistryPrefix}/${IMG_NAME}
docker push ${QiniuRegistryPrefix}/${IMG_NAME}
