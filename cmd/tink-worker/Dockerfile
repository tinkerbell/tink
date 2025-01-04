FROM alpine:3.20.3

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update --upgrade ca-certificates=20241121-r0

COPY bin/tink-worker-${TARGETOS}-${TARGETARCH} /usr/bin/tink-worker

ENTRYPOINT [ "/usr/bin/tink-worker" ]
