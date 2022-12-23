FROM alpine:3.15

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update --upgrade ca-certificates

COPY bin/tink-worker-${TARGETOS}-${TARGETARCH} /usr/bin/tink-worker

ENTRYPOINT [ "/usr/bin/tink-worker" ]
