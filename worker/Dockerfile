
# TODO: update the dockerfile

# FROM alpine:3.10
FROM quickdevnotes/docker

EXPOSE 4444
# RUN apk add --no-cache --update --upgrade ca-certificates
COPY rover-worker-linux-x86_64 /bin/worker
ENTRYPOINT [ "/bin/worker" ]