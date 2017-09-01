#!/bin/sh

set -e

cd /go/src/github.com/kayac/alphawing

git fetch
git checkout $GIT_BRANCH
git pull

mkdir -p /tmp/alphawing

revel build github.com/kayac/alphawing /tmp/alphawing

go build -a -tags netgo -installsuffix netgo app/tmp/main.go
