FROM golang:1.13-alpine

EXPOSE 42113
EXPOSE 42114

WORKDIR /go/src/app

COPY . .

RUN apk update && \
	apk add ca-certificates postgresql-client && \
	apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl && \
	go build -o /go/bin/tink-server .

COPY deploy/migrate /migrate
COPY deploy/docker-entrypoint-initdb.d/tinkerbell-init.sql /init.sql

ENTRYPOINT ["tink-server"]
