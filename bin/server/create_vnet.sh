#!/bin/bash

docker network create -d bridge --attachable --subnet 10.9.1.0/24 --gateway 10.9.1.1 nfs_net


