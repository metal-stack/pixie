#!/bin/sh

set -x
set -e

mkdir -p /tmp/go/src/go.universe.tf
if [ -d /tmp/stuff/.git ]; then
    echo "Building from local dev copy"
    mkdir -p /tmp/go/src/go.universe.tf
    cp -R /tmp/stuff /tmp/go/src/go.universe.tf/netboot
else
    echo "Building from git checkout"
fi

export GOPATH=/tmp/go
echo "http://dl-4.alpinelinux.org/alpine/edge/community" >>/etc/apk/repositories
apk -U add ca-certificates git go gcc musl-dev
apk upgrade
go get -v github.com/Masterminds/glide
go get -v -d go.universe.tf/netboot/cmd/pixiecore
cd /tmp/go/src/go.universe.tf/netboot
/tmp/go/bin/glide install
cd cmd/pixiecore
go build .
cp ./pixiecore /pixiecore
cd /
apk del git go gcc musl-dev
rm -rf /tmp/go /tmp/stuff /root/.glide /usr/lib/go /var/cache/apk/*
