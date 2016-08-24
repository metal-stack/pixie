FROM alpine:edge
MAINTAINER David Anderson <dave@natulte.net>

COPY . /tmp/go/src/go.universe.tf/netboot
RUN \
  export GOPATH=/tmp/go &&\
  echo "http://dl-4.alpinelinux.org/alpine/edge/community" >>/etc/apk/repositories &&\
  apk -U add ca-certificates git go gcc musl-dev &&\
  apk upgrade &&\
  go get go.universe.tf/netboot/cmd/pixiecore &&\
  cp /tmp/go/bin/pixiecore /pixiecore &&\
  rm -rf /tmp/go &&\
  apk del git go gcc musl-dev &&\
  rm -rf /var/cache/apk/*

ENTRYPOINT ["/pixiecore"]
