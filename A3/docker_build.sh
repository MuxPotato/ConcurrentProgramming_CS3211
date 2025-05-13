#!/bin/bash
if [ $# -ne 1 ]; then
    echo "Usage: ./docker_build.sh <img_name>"
else
    docker build -t $1 .
fi;