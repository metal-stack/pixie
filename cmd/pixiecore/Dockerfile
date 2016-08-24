FROM alpine:edge
MAINTAINER David Anderson <dave@natulte.net>

COPY . /tmp/go/src/go.universe.tf/netboot
RUN \
  export GOPATH=/tmp/go &&\
  echo "http://dl-4.alpinelinux.org/alpine/edge/community" >>/etc/apk/repositories &&\
  apk -U add ca-certificates git go gcc musl-dev &&\
  apk upgrade &&\
  cd /tmp/go/src/go.universe.tf/netboot &&\
  go get -v . &&\
  cd cmd/pixiecore &&\
  go build . &&\
  cp pixiecore /pixiecore &&\
  cd / &&\
  rm -rf /tmp/go &&\
  apk del git go gcc musl-dev &&\
  rm -rf /var/cache/apk/*

ENTRYPOINT ["/pixiecore"]
