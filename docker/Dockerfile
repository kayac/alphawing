FROM golang:alpine AS build-env
MAINTAINER mix3

RUN apk --update add wget ca-certificates unzip git gcc musl-dev

ADD patch         /tmp/patch
ADD docker_run.sh /tmp/docker_run.sh

RUN mkdir -p /go/src/code.google.com/p \
 && cd /go/src/code.google.com/p \
 && wget -O /tmp/z.$$ https://storage.googleapis.com/google-code-archive-source/v2/code.google.com/go-uuid/source-archive.zip              && unzip /tmp/z.$$ && rm /tmp/z.$$ \
 && wget -O /tmp/z.$$ https://storage.googleapis.com/google-code-archive-source/v2/code.google.com/goauth2/source-archive.zip              && unzip /tmp/z.$$ && rm /tmp/z.$$ \
 && wget -O /tmp/z.$$ https://storage.googleapis.com/google-code-archive-source/v2/code.google.com/google-api-go-client/source-archive.zip && unzip /tmp/z.$$ && rm /tmp/z.$$

RUN go get github.com/mattn/go-sqlite3 \
 && go get github.com/revel/revel \
 && go get github.com/revel/cmd/revel \
 && go get github.com/coopernurse/gorp \
 && go get github.com/DHowett/go-plist \
 && go get github.com/google/go-github/github \
 && go get github.com/shogo82148/androidbinary \
 && go get github.com/go-sql-driver/mysql \
 && go get cloud.google.com/go/storage \
 && cd /go/src/github.com/revel/cmd && git checkout v0.17 && patch -p1 < /tmp/patch && go install github.com/revel/cmd/revel \
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

FROM golang:alpine
COPY --from=build-env /go/src/github.com/kayac/alphawing/alphawing.tar.gz /tmp/alphawing.tar.gz
RUN tar xvzf /tmp/alphawing.tar.gz -C /go
CMD sh /go/run.sh
