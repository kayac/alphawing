## build

```
$ docker build [--build-arg "GIT_BRANCH=master"] [--build-arg "GIT_REMOTE=https://github.com/kayac/alphawing"] -t alphawing:latest --no-cache docker/
```

## run

### /path/to/app.conf

```
db.spec                       = /data/alphawing.db
google.serviceaccount.keypath = /go/src/github.com/kayac/alphawing/conf/key.json
```

```
$ docker run --rm \
    -v /path/to/data:/data \
    -v /path/to/app.conf:/go/src/github.com/kayac/alphawing/conf/app.conf \
    -v /path/to/key.json:/go/src/github.com/kayac/alphawing/conf/key.json \
    -p 9000:9000 \
    -it alphawing:latest
```
