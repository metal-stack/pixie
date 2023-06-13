FROM golang:1.20-bullseye as builder
WORKDIR /work
COPY . .
RUN apt update \
 && apt install --yes --no-install-recommends \
    liblzma-dev \
 && make ipxe pixie

FROM alpine:3.18
RUN apk -U add ca-certificates
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
