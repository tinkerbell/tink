FROM alpine:3.7

CMD sleep 60d

RUN apk add --no-cache --update --upgrade ca-certificates
COPY rover-linux-x86_64 /bin/rover
COPY sample.tmpl /tmp
