#!/bin/sh
export GOPATH
GOPATH="`pwd`"
cd src/vouquet
dep ensure
dep status

