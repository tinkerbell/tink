FROM alpine:3.11

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tink-worker /tink-worker
ENTRYPOINT [ "/tink-worker" ]
