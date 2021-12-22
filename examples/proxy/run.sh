#!/bin/bash

set -e
GOOS=linux GOARCh=amd64 go build -o proxy
docker run --rm -ti --net nfs_net --name proxy --platform linux/amd64 -v `pwd`:/opt/nfs -w /opt/nfs centos:7 ./proxy
rm -f ./proxy
