FROM golang:1.19-buster as builder
WORKDIR /work
COPY . .
RUN apt update \
 && apt install --yes --no-install-recommends \
    liblzma-dev \
 && make update-ipxe pixie

FROM alpine:3.16
RUN apk -U add ca-certificates
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
