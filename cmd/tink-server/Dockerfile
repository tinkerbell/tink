FROM alpine:3.11

ENTRYPOINT ["tink-server"]
EXPOSE 42113
EXPOSE 42114

RUN apk add --update ca-certificates && \
    apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl

COPY deploy/db/migrate /migrate
COPY deploy/db/docker-entrypoint-initdb.d/tinkerbell-init.sql /init.sql
COPY tink-server /bin/
