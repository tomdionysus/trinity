#!/bin/sh
# This script launch several node so it easier to test federation and data repartition

build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11212 --node localhost:13531 --port 13532 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11213 --node localhost:13531 --port 13533 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11214 --node localhost:13531 --port 13534 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11215 --node localhost:13531 --port 13535 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11216 --node localhost:13531 --port 13536 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11217 --node localhost:13531 --port 13537 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11218 --node localhost:13531 --port 13538 &
build/trinity-server --ca cert/ca.pem --cert cert/localhost.pem --loglevel info -memcache --memcacheport 11219 --node localhost:13531 --port 13539 &
