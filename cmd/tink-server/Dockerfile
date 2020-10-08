FROM golang:1.15-alpine
RUN apk add --no-cache --update --upgrade ca-certificates make git
COPY . /usr/myapp
WORKDIR /usr/myapp
RUN make server

FROM alpine:3.11
ENTRYPOINT ["/usr/bin/tink-server"]
EXPOSE 42113
EXPOSE 42114

RUN apk add --update ca-certificates && \
    apk add --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing cfssl
COPY --from=0 /usr/myapp/cmd/tink-server/tink-server /usr/bin/tink-server
