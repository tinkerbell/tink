FROM alpine:3.10

RUN apk add --no-cache --update --upgrade ca-certificates
COPY rover-worker-linux-x86_64 /bin/worker
ENTRYPOINT [ "/bin/worker" ]