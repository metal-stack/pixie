FROM golang:1.11-stretch as builder
RUN apt update \
 && apt-get install make
COPY . /go/src/go.universe.tf/netboot
WORKDIR /go/src/go.universe.tf/netboot
RUN make ci-prepare
RUN cd cmd/pixiecore \
 && CGO_ENABLE=0 go build -tags netgo

FROM alpine:latest
LABEL maintainer FI-TS Devops <devops@f-i-ts.de>
RUN apk -U add ca-certificates
COPY --from=builder /go/src/go.universe.tf/netboot/cmd/pixiecore/pixiecore /pixiecore
ENTRYPOINT ["/pixiecore"]
