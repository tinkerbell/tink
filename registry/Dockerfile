FROM alpine:3.10 AS base
RUN apk add --no-cache --update --upgrade ca-certificates
RUN apk add --no-cache --update --upgrade --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY tls /tls
WORKDIR /tls
RUN [ "sh", "gencerts.sh" ]


FROM registry:2 AS auth
ARG REGISTRY_USERNAME
ARG REGISTRY_PASSWORD
RUN mkdir /auth \
  && htpasswd -Bbn ${REGISTRY_USERNAME} ${REGISTRY_PASSWORD} > /auth/htpasswd

FROM registry:2
RUN mkdir -p /certs /auth 
COPY --from=base /tls/server.pem /certs
COPY --from=base /tls/server-key.pem /certs
COPY --from=auth /auth/htpasswd /auth
EXPOSE 443

