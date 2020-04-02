FROM alpine:3.7

CMD sleep 60d

RUN apk add --no-cache --update --upgrade ca-certificates
COPY tink-cli /bin/tink
COPY sample.tmpl /tmp
