FROM golang:1.15-alpine
RUN apk add --no-cache --update --upgrade ca-certificates make git
COPY . /usr/myapp
WORKDIR /usr/myapp
RUN make cli

FROM alpine:3.11
CMD sleep infinity

RUN apk add --no-cache --update --upgrade ca-certificates
COPY --from=0 /usr/myapp/cmd/tink-cli/tink-cli /usr/bin/tink
