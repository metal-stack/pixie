FROM alpine:edge
MAINTAINER David Anderson <dave@natulte.net>

COPY . /tmp/stuff
RUN cd /tmp/stuff; [ -f build.sh ] && ./build.sh || ./dockerfiles/pixiecore/build.sh

ENTRYPOINT ["/pixiecore"]
