FROM alpine:3.15
CMD sleep infinity

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tink-cli-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink
