FROM alpine:3.11

CMD sleep infinity

RUN apk add --no-cache --update --upgrade ca-certificates

COPY tink-cli /bin/tink
COPY sample.tmpl /tmp
