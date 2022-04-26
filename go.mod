module github.com/tinkerbell/tink

go 1.13

require (
	github.com/bufbuild/buf v1.0.0-rc2
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/equinix-labs/otel-init-go v0.0.1
	github.com/go-logr/zapr v1.2.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/ktr0731/evans v0.10.0
	github.com/lib/pq v1.10.1
	github.com/matryer/moq v0.2.3
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.18.1
	github.com/opencontainers/image-spec v1.0.2
	github.com/packethost/pkg v0.0.0-20200903155310-0433e0605550
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0
	go.uber.org/multierr v1.7.0
	google.golang.org/genproto v0.0.0-20211021150943-2b146023228c
	google.golang.org/grpc v1.42.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	knative.dev/pkg v0.0.0-20211119170723-a99300deff34
	mvdan.cc/gofumpt v0.1.1
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/controller-runtime/tools/setup-envtest v0.0.0-20220304125252-9ee63fc65a97
	sigs.k8s.io/controller-tools v0.8.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/go-openapi/strfmt v0.19.3 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	go.mongodb.org/mongo-driver v1.1.2 // indirect
)

replace github.com/stormcat24/protodep => github.com/ackintosh/protodep v0.0.0-20200728152107-abf8eb579d6c

// 2.8.0+incompatible has incorrect checksums in the public sumdb.
exclude github.com/docker/distributions v2.8.0+incompatible
