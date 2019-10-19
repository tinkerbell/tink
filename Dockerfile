FROM alpine:3.7

ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "/rover" ]
EXPOSE 42111
EXPOSE 42112

RUN apk add --no-cache --update --upgrade ca-certificates postgresql-client
RUN apk add --no-cache --update --upgrade --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY entrypoint.sh /entrypoint.sh
COPY tls /tls
COPY deploy/migrate /migrate
COPY deploy/docker-entrypoint-initdb.d/rover-init.sql /init.sql
COPY rover-linux-x86_64 /rover
