#!/usr/bin/env bash
#
# protoc.sh uses the local protoc if installed, otherwise
# docker will be used with a complete environment provided
# by https://github.com/jaegertracing/docker-protobuf.
# Alternative images like grpc/go are very dated and do not
# include the needed plugins and includes.
#
set -e

GOPATH=${GOPATH:-$(go env GOPATH)}

if command -v protoc >/dev/null; then
	GW_PATH="$GOPATH"/src/github.com/grpc-ecosystem/grpc-gateway
	GO111MODULES=on go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

	PROTOC="protoc -I/usr/local/include -I$GW_PATH/third_party/googleapis --plugin=protoc-gen-grpc-gateway=$GOPATH/bin/protoc-gen-grpc-gateway"
else
	IMAGE=jaegertracing/protobuf:0.2.0
	BASE=/protos
	PROTOC="docker run -v $(pwd):$BASE -w $BASE --rm $IMAGE "
fi

for proto in hardware packet template workflow; do
	echo "Generating ${proto}.pb.go..."
	$PROTOC -I./ \
		-I./common \
		--go_opt=paths=source_relative \
		--go_out=plugins=grpc:./ \
		--grpc-gateway_out=logtostderr=true:. \
		"${proto}/${proto}.proto"
done
goimports -w .
