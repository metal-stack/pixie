FROM golang:1.11-stretch as builder
RUN apt update
RUN CGO_ENABLE=0 go build -tags netgo -o /pixiecore

FROM alpine:latest
LABEL maintainer FI-TS Devops <devops@f-i-ts.de>
RUN apk -U add ca-certificates
COPY --from=builder /pixiecore /pixiecore
ENTRYPOINT ["/pixiecore"]
