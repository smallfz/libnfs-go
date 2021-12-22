#!/bin/bash

set -e
GOOS=linux GOARCH=amd64 go build -o server_linux
docker run --rm -ti --net nfs_net --name nfs-server -p 2049:2049 --platform linux/amd64 -v `pwd`:/opt/nfs -w /opt/nfs centos:7 ./server_linux
