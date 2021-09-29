module github.com/tinkerbell/tink

go 1.13

require (
	github.com/bufbuild/buf v1.0.0-rc2
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/equinix-labs/otel-init-go v0.0.1
	github.com/go-openapi/strfmt v0.19.3 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/ktr0731/evans v0.10.0
	github.com/lib/pq v1.10.1
	github.com/matryer/moq v0.2.3
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/packethost/pkg v0.0.0-20200903155310-0433e0605550
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.1
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	google.golang.org/genproto v0.0.0-20210921142501-181ce0d877f6
	google.golang.org/grpc v1.41.0-dev.0.20210907181116-2f3355d2244e
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.22.2
	mvdan.cc/gofumpt v0.1.1
	sigs.k8s.io/controller-runtime v0.10.1
)

replace github.com/stormcat24/protodep => github.com/ackintosh/protodep v0.0.0-20200728152107-abf8eb579d6c
