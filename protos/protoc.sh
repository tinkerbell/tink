#!/bin/bash

set -e

for proto in $(echo hardware target template workflow); do
	echo "Generating ${proto}.pb.go..."
	protoc -I ./ -I ./common/ ${proto}/${proto}.proto --go_out=plugins=grpc:./
done
