#!/bin/bash

docker run --rm -ti --net nfs_net --platform=linux/amd64/v8 --privileged -v `pwd`:/data centos:18-nfs-cl /bin/bash
