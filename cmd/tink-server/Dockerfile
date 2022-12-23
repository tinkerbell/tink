FROM alpine:3.15

ARG TARGETOS
ARG TARGETARCH

EXPOSE 42113 42114

RUN apk add --no-cache --update --upgrade ca-certificates

COPY bin/tink-server-${TARGETOS}-${TARGETARCH} /usr/bin/tink-server

ENTRYPOINT ["/usr/bin/tink-server"]
