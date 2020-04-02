FROM alpine:3.10

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tink-worker /tink-worker
ENTRYPOINT [ "/tink-worker" ]
