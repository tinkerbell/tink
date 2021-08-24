module github.com/tinkerbell/tink

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-openapi/strfmt v0.19.3 // indirect
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.15.2
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/lib/pq v1.2.1-0.20191011153232-f91d3411e481
	github.com/mattn/go-runewidth v0.0.5 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/packethost/pkg v0.0.0-20200903155310-0433e0605550
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.1-0.20200713175500-884edc58ad08
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stormcat24/protodep v0.0.0-20200505140716-b02c9ba62816
	github.com/stretchr/testify v1.6.1
	github.com/testcontainers/testcontainers-go v0.9.0
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20201026171402-d4b8fe4fd877
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)

replace github.com/stormcat24/protodep => github.com/ackintosh/protodep v0.0.0-20200728152107-abf8eb579d6c
