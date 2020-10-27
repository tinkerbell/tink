#!/usr/bin/env bash
#
# protoc.sh uses the local protoc if installed, otherwise
# docker will be used with a complete environment provided
# by https://github.com/jaegertracing/docker-protobuf.
# Alternative images like grpc/go are very dated and do not
# include the needed plugins and includes.
#
set -e

export GOBIN=$PWD/bin

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null 2>&1 && pwd)"
export GOBIN=$DIR/bin
export PATH=$GOBIN:$PATH
unset DIR

# shellcheck disable=SC2046
go install $(sed -n -e 's|^\s*_\s*"\(.*\)".*$|\1| p' tools.go)

protodep up -f --use-https

for proto in protos/*/*.proto; do
	echo "Generating ${proto/.proto/}.pb.go..."
	protoc \
		-I./protos \
		-I./protos/third_party/ \
		--go_out ./protos \
		--go_opt paths=source_relative \
		--go_opt plugins=grpc \
		--grpc-gateway_out ./protos \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		"${proto}"
done
goimports -w .
