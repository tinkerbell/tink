FROM alpine:3.11
CMD sleep infinity

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache --update --upgrade ca-certificates
COPY cmd/tink-cli/tink-cli-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink
