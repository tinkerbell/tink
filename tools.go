// +build tools

package tools

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/matryer/moq"
	_ "github.com/stormcat24/protodep"
	_ "mvdan.cc/gofumpt"
)
