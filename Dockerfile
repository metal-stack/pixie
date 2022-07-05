FROM golang:1.18-buster as builder
WORKDIR /work
COPY . .
RUN make

FROM alpine:3.16
RUN apk -U add ca-certificates
COPY --from=builder /work/build/pixie /pixie
ENTRYPOINT ["/pixie"]
