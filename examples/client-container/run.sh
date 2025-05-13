#!/bin/bash

docker run --rm -ti --net nfs_net --platform=linux/amd64/v8 --privileged -v `pwd`:/data nfs-test:latest /bin/bash
