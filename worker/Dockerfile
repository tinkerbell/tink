FROM alpine:3.10

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tinkerbell-worker /bin/worker
ENTRYPOINT [ "/bin/worker" ]
