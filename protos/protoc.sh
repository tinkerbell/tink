#!/usr/bin/env bash

set -e

for proto in hardware template workflow; do
	echo "Generating ${proto}.pb.go..."
	protoc -I ./ -I ./common/ "${proto}/${proto}.proto" --go_out=plugins=grpc:./
done
