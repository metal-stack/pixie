FROM debian:13 AS ipxe-builder
WORKDIR /work
COPY . .
RUN apt update \
 && apt install --yes --no-install-recommends \
    ca-certificates \
    gcc \
    git \
    libc6-dev \
    liblzma-dev \
    make \
 && make ipxe

FROM golang:1.27-trixie AS builder
WORKDIR /work
COPY . .
COPY --from=ipxe-builder /work/ipxe/ipxe /work/ipxe/ipxe
RUN make pixie

FROM gcr.io/distroless/static
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
