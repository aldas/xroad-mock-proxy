#!/usr/bin/env bash

SERVER=${1:-http://localhost:8081}
PAYLOAD=${2:-test/testdata/api/proxy/add-server.json}

if [[ ! -f ${PAYLOAD} ]]; then
    echo "File not found! ${PAYLOAD}"
    exit 1
fi

curl -X POST -d @${PAYLOAD} --header "Content-Type: application/json;charset=UTF-8" ${SERVER}/api/servers
