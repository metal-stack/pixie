# Pixiecore

This project is a permanent fork of: [Pixiecore](https://github.com/danderson/netboot/tree/master/pixiecore)

Sample command to run `pixie` in grpc mode which talks to the metal-api and provides grpc client certificates and the metal-api-view-hmac to the metal-hammer.
With this metal-hammer will be able to talk to metal-api directly.

```bash
docker run -it --rm -name pixiecore \
    --network host \
    --dns 10.1.253.13 \
    --dns 10.1.253.29 \
    --volume "/certs/grpc:/certs/grpc:ro" \
    ghcr.io/metal-stack/pixie grpc \
        --debug \
        --dhcp-no-bind \
        --pixie-api-url http://the-ip-of-this-service/certs \
        --grpc-address api.metal-stack.dev:50051 \
        --grpc-ca-cert /certs/grpc/ca.pem \
        --grpc-cert /certs/grpc/client.pem \
        --grpc-key /certs/grpc/client-key.pem \
        --metal-api-url https://api.metal-stack.io/metal \
        --metal-api-view-hmac a-view-hmac \
        --partition partition-1
        --ntp-servers 0.custom.ntp,1.custom.ntp
```
