FROM alpine:3.15

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update --upgrade ca-certificates

COPY bin/virtual-worker-${TARGETOS}-${TARGETARCH} /usr/bin/virtual-worker

ENTRYPOINT [ "/usr/bin/virtual-worker" ]
