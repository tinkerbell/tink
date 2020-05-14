FROM alpine:3.11

ENTRYPOINT ["tink-server"]
EXPOSE 42113
EXPOSE 42114

RUN apk add --update ca-certificates && \
    apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl

COPY tink-server /bin/
