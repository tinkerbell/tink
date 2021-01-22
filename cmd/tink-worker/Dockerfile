FROM alpine:3.11
ENTRYPOINT [ "/usr/bin/tink-worker" ]

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache --update --upgrade ca-certificates
COPY cmd/tink-worker/tink-worker-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink-worker
