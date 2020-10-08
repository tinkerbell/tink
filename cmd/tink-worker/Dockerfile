FROM golang:1.15-alpine
RUN apk add --no-cache --update --upgrade ca-certificates make git
COPY . /usr/myapp
WORKDIR /usr/myapp
RUN make worker

FROM alpine:3.11
ENTRYPOINT [ "/usr/bin/tink-worker" ]
RUN apk add --no-cache --update --upgrade ca-certificates
COPY --from=0 /usr/myapp/cmd/tink-worker/tink-worker /usr/bin/tink-worker
