#!/bin/sh

set -e

cd /go/src/github.com/kayac/alphawing

if [ "$GIT_REMOTE" != "" ]; then
    git remote rename origin temp
    git remote add origin $GIT_REMOTE
fi

git fetch origin
git checkout $GIT_BRANCH
git pull

revel package github.com/kayac/alphawing
