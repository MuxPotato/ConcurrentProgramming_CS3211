#!/bin/bash
if [ $# -ne 5 ]; then
    echo "Usage: ./docker_run.sh <img_name> <port> <seed> <clients> <messages>"
else 
    docker run --rm \
        --ulimit nofile=1000000:1000000 \
        $1 /usr/local/bin/app "$2" "$3" "$4" "$5"
fi;