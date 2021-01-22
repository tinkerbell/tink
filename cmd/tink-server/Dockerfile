FROM alpine:3.11
ENTRYPOINT ["/usr/bin/tink-server"]
EXPOSE 42113
EXPOSE 42114

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --update ca-certificates && \
    apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY cmd/tink-server/tink-server-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink-server
