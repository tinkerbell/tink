ARG REGISTRY
FROM $REGISTRY/ubuntu:base
COPY grub.default /tmp/
COPY cleanup.sh /
ENTRYPOINT ["./cleanup.sh"]

