FROM alpine:3.11
ENTRYPOINT [ "/entrypoint.sh" ]

RUN apk add --no-cache --update --upgrade ca-certificates postgresql-client
RUN apk add --no-cache --update --upgrade --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl

COPY . .
