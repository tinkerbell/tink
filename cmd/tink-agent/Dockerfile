FROM alpine:3.15

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update --upgrade ca-certificates

COPY bin/tink-agent-${TARGETOS}-${TARGETARCH} /usr/bin/tink-agent

ENTRYPOINT ["/usr/bin/tink-agent"]
