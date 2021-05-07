// +build tools

package tools

import (
	_ "google.golang.org/protobuf/protoc-gen-go"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/stormcat24/protodep"
)
