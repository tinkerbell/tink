FROM alpine:3.20.3

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update --upgrade ca-certificates=20241121-r0

COPY bin/tink-agent-${TARGETOS}-${TARGETARCH} /usr/bin/tink-agent

ENTRYPOINT ["/usr/bin/tink-agent"]
