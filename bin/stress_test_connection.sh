#!/bin/sh
# This script just spawn and kill instance to see how the system react

while [ 1 ]; do
    timeout 1 build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info --node localhost:13531 --port 13539
done