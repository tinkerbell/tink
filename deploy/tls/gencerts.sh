#!/usr/bin/env sh

set -eux

cd /certs

if [ ! -f ca-key.pem ]; then
	cfssl gencert \
		-initca ca.json | cfssljson -bare ca
fi

if [ ! -f server.pem ]; then
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=/ca-config.json \
		-profile=server \
		server-csr.json |
		cfssljson -bare server
fi

cat server.pem ca.pem >bundle.pem.tmp

# only "modify" the file if truly necessary since workflow will serve it with
# modtime info for client caching purposes
if ! cmp -s bundle.pem.tmp bundle.pem; then
	mv bundle.pem.tmp bundle.pem
else
	rm bundle.pem.tmp
fi
