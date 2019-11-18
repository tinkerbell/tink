FROM registry:2 AS auth
ARG REGISTRY_USERNAME
ARG REGISTRY_PASSWORD
RUN mkdir /auth \
  && htpasswd -Bbn ${REGISTRY_USERNAME} ${REGISTRY_PASSWORD} > /auth/htpasswd

FROM registry:2
RUN mkdir -p /certs /auth 
COPY --from=auth /auth/htpasswd /auth
EXPOSE 443

