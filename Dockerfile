FROM golang:1.13-buster as builder
RUN apt update \
 && apt-get install make
COPY . /work/
WORKDIR /work
RUN cd cmd/pixiecore \
 && CGO_ENABLE=0 go build -trimpath -tags netgo \
 && strip /work/cmd/pixiecore/pixiecore

FROM alpine:3.11
RUN apk -U add ca-certificates
COPY --from=builder /work/cmd/pixiecore/pixiecore /pixiecore
ENTRYPOINT ["/pixiecore"]
