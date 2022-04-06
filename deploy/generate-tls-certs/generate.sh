#!/usr/bin/env bash

cfssl gencert -initca /code/tls/csr.json | cfssljson -bare ca -
cfssl gencert -config /code/tls/ca-config.json -ca ca.pem -ca-key ca-key.pem -profile server /code/tls/csr.json | cfssljson -bare server
cat server.pem ca.pem >/certs/"${FACILITY:-onprem}"/bundle.pem
mv server-key.pem /certs/"${FACILITY:-onprem}"/server-key.pem
rm -rf ca-key.pem ca.csr ca.pem server.csr server.pem
