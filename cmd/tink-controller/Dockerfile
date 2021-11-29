FROM alpine:3.15
ENTRYPOINT ["/usr/bin/tink-controller"]
EXPOSE 42113
EXPOSE 42114

ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --update ca-certificates
COPY tink-controller-linux-${TARGETARCH:-amd64}${TARGETVARIANT} /usr/bin/tink-controller
