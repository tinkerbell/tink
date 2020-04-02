FROM alpine:3.11

ENTRYPOINT [ "/tink-worker" ]

RUN apk add --no-cache --update --upgrade ca-certificates

COPY tink-worker /tink-worker
