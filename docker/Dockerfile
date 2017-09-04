FROM golang:alpine AS build-env
MAINTAINER mix3

RUN apk --update add wget ca-certificates unzip git gcc musl-dev

ADD docker_run.sh /tmp/docker_run.sh

RUN mkdir -p /go/src/code.google.com/p
COPY goauth2              /go/src/code.google.com/p/goauth2
COPY go-uuid              /go/src/code.google.com/p/go-uuid
COPY google-api-go-client /go/src/code.google.com/p/google-api-go-client

RUN go get github.com/revel/cmd/revel \
 && go get github.com/DHowett/go-plist \
 && go get github.com/coopernurse/gorp \
 && go get github.com/go-sql-driver/mysql \
 && go get github.com/google/go-github/github \
 && go get github.com/mattn/go-sqlite3 \
 && go get github.com/pborman/uuid \
 && go get github.com/shogo82148/androidbinary \
 && go get cloud.google.com/go/storage \
 && cd /go/src/github.com/revel/cmd   && git checkout v0.17 \
 && cd /go/src/github.com/revel/revel && git checkout v0.17.1

RUN mkdir -p /go/src/github.com/kayac \
 && cd /go/src/github.com/kayac \
 && git clone https://github.com/kayac/alphawing.git \
 && cd alphawing \
 && cp conf/app.conf.sample conf/app.conf

ARG GIT_BRANCH=master
ENV GIT_BRANCH $GIT_BRANCH
ARG GIT_REMOTE
ENV GIT_REMOTE $GIT_REMOTE
RUN sh /tmp/docker_run.sh

FROM alpine:latest
RUN apk --update add ca-certificates
COPY --from=build-env /go/src/github.com/kayac/alphawing/alphawing.tar.gz /tmp/alphawing.tar.gz
RUN mkdir /go
RUN tar xvzf /tmp/alphawing.tar.gz -C /go
CMD sh /go/run.sh
