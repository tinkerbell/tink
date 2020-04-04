ARG REGISTRY
FROM $REGISTRY/ubuntu:base
COPY repos.sh /
COPY grub.default /
COPY grub-installer.sh /
COPY install.sh /
ENTRYPOINT ["./install.sh"]

