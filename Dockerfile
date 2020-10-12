FROM golang:1.15-buster as builder
COPY . /work/
WORKDIR /work
RUN cd cmd/pixiecore \
 && CGO_ENABLE=0 go build -trimpath -tags netgo \
 && strip /work/cmd/pixiecore/pixiecore

FROM alpine:3.12
RUN apk -U add ca-certificates
COPY --from=builder /work/cmd/pixiecore/pixiecore /pixiecore
ENTRYPOINT ["/pixiecore"]
