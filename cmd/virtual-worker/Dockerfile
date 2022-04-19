FROM alpine:3.15
ENTRYPOINT [ "/usr/bin/virtual-worker" ]

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache --update --upgrade ca-certificates
COPY virtual-worker-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/virtual-worker
