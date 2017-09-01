FROM golang:alpine
MAINTAINER mix3

RUN apk --update add wget ca-certificates unzip git gcc musl-dev

ADD patch         /tmp/patch
ADD docker_run.sh /tmp/docker_run.sh

RUN mkdir -p /go/src/code.google.com/p
RUN cd /go/src/code.google.com/p \
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

RUN mkdir -p /go/src/github.com/kayac
RUN cd /go/src/github.com/kayac && git clone https://github.com/kayac/alphawing.git
ADD app.conf /go/src/github.com/kayac/alphawing/conf/app.conf
CMD sh /tmp/docker_run.sh
