# be aware that bookworm has a newer gcc which can not compile the older ipxe
FROM golang:1.22-bullseye as builder
WORKDIR /work
COPY . .
RUN apt update \
 && apt install --yes --no-install-recommends \
    liblzma-dev \
 && make ipxe pixie

FROM alpine:3.19
RUN apk -U add ca-certificates
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
