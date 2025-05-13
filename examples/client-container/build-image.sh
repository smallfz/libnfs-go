#!/bin/bash

set -e
docker build -t nfs-test:latest --platform linux/amd64 .

