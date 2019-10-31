FROM alpine:3.10 AS base
RUN apk add --no-cache --update --upgrade ca-certificates
RUN apk add --no-cache --update --upgrade --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY tls /tls
WORKDIR /tls
RUN [ "sh", "gencerts.sh" ]

FROM registry:2
WORKDIR /certs
COPY --from=base /tls/server.pem .
COPY --from=base /tls/server-key.pem .
ENV REGISTRY_HTTP_ADDR=0.0.0.0:443
ENV REGISTRY_HTTP_TLS_CERTIFICATE=/certs/server.pem
ENV REGISTRY_HTTP_TLS_KEY=/certs/server-key.pem
EXPOSE 443

