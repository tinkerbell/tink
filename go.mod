module github.com/tinkerbell/tink

go 1.13

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/docker/go-units v0.4.0 // indirect
	github.com/equinix-labs/otel-init-go v0.0.1
	github.com/go-openapi/strfmt v0.19.3 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/lib/pq v1.2.1-0.20191011153232-f91d3411e481
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/packethost/pkg v0.0.0-20200903155310-0433e0605550
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stormcat24/protodep v0.0.0-20200505140716-b02c9ba62816
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.9.0
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/stormcat24/protodep => github.com/ackintosh/protodep v0.0.0-20200728152107-abf8eb579d6c
