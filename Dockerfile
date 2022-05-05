FROM golang:1.18-buster as builder
COPY . /work/
WORKDIR /work
RUN make

FROM alpine:3.15
RUN apk -U add ca-certificates
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
