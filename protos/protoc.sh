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

if command -v protoc >/dev/null; then
	GW_PATH="${GOPATH:-$(go env GOPATH)}/src/github.com/grpc-ecosystem/grpc-gateway"

	DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null 2>&1 && pwd)"
	export GOBIN=$DIR/bin
	export PATH=$GOBIN:$PATH

	# shellcheck disable=SC2046
	go install $(sed -n -e 's|^\s*_\s*"\(.*\)".*$|\1| p' "$DIR/tools.go")

	PROTOC="protoc -I$GW_PATH/third_party/googleapis"

	unset DIR GW_PATH
else
	IMAGE=jaegertracing/protobuf:0.2.0
	BASE=/protos
	PROTOC="docker run -v $(pwd):$BASE -w $BASE --rm $IMAGE "
fi

for proto in hardware packet template workflow; do
	echo "Generating ${proto}.pb.go..."
	$PROTOC \
		-I./ \
		-I./common \
		--go_out ./protos \
		--go_opt paths=source_relative \
		--go_opt plugins=grpc \
		--grpc-gateway_out ./protos \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		"${proto}/${proto}.proto"
done
goimports -w .
