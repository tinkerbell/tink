FROM alpine:3.7

CMD sleep 60d

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tinkerbell-cli /bin/tinkerbell
COPY sample.tmpl /tmp
