#!/usr/bin/env bash

# usage: ./post2mock.sh testdata/rr.rr456.v1/rr456.paring.xml

PAYLOAD=${1:-testdata/rr.rr456.v1/rr456.paring.xml}
SERVER=${2:-http://localhost:8082}

if [[ ! -f ${PAYLOAD} ]]; then
    echo "File not found!"
    exit 1
fi

curl -X POST -d @${PAYLOAD} --header "Content-Type: text/xml;charset=UTF-8" ${SERVER}/cgi-bin/consumer_proxy
