FROM alpine:3.7

ENTRYPOINT [ "/rover" ]
EXPOSE 42113
EXPOSE 42114

RUN apk add --no-cache --update --upgrade ca-certificates postgresql-client
RUN apk add --no-cache --update --upgrade --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY deploy/migrate /migrate
COPY deploy/docker-entrypoint-initdb.d/rover-init.sql /init.sql
COPY rover-server /rover
