# be aware that bookworm has a newer gcc which can not compile the older ipxe
FROM golang:1.24-bullseye AS builder
WORKDIR /work
COPY . .
RUN apt update \
 && apt install --yes --no-install-recommends \
    liblzma-dev \
 && make ipxe pixie

FROM gcr.io/distroless/static-debian12
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
