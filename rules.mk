# Only use the recipes defined in these makefiles
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:
# Delete target files if there's an error
# This avoids a failure to then skip building on next run if the output is created by shell redirection for example
# Not really necessary for now, but just good to have already if it becomes necessary later.
.DELETE_ON_ERROR:
# Treat the whole recipe as a one shell script/invocation instead of one-per-line
.ONESHELL:
# Use bash instead of plain sh
SHELL := bash
.SHELLFLAGS := -o pipefail -euc

binaries := cmd/tink-cli/tink-cli cmd/tink-controller/tink-controller cmd/tink-server/tink-server cmd/tink-worker/tink-worker cmd/virtual-worker/virtual-worker
version := $(shell git rev-parse --short HEAD)
tag := $(shell git tag --points-at HEAD)
ifneq (,$(tag))
version := $(tag)-$(version)
endif
LDFLAGS := -ldflags "-X main.version=$(version)"
export CGO_ENABLED := 0

.PHONY: server cli worker virtual-worker test $(binaries)
cli: cmd/tink-cli/tink-cli
controller: cmd/tink-controller/tink-controller
server: cmd/tink-server/tink-server
worker : cmd/tink-worker/tink-worker
virtual-worker : cmd/virtual-worker/virtual-worker

crossbinaries := $(addsuffix -linux-,$(binaries))
crossbinaries := $(crossbinaries:=386) $(crossbinaries:=amd64) $(crossbinaries:=arm64) $(crossbinaries:=armv6) $(crossbinaries:=armv7)

.PHONY: crosscompile $(crossbinaries)
%-386:   FLAGS=GOOS=linux GOARCH=386
%-amd64: FLAGS=GOOS=linux GOARCH=amd64
%-arm64: FLAGS=GOOS=linux GOARCH=arm64
%-armv6: FLAGS=GOOS=linux GOARCH=arm GOARM=6
%-armv7: FLAGS=GOOS=linux GOARCH=arm GOARM=7
$(binaries) $(crossbinaries):
	$(FLAGS) go build $(LDFLAGS) -o $@ ./$(@D)

.PHONY: tink-cli-image tink-controller-image tink-server-image tink-worker-image virtual-worker-image
tink-cli-image: cmd/tink-cli/tink-cli-linux-amd64
	docker build -t tink-cli cmd/tink-cli/
tink-controller-image: cmd/tink-controller/tink-controller-linux-amd64
	docker build -t tink-controller cmd/tink-controller/
tink-server-image: cmd/tink-server/tink-server-linux-amd64
	docker build -t tink-server cmd/tink-server/
tink-worker-image: cmd/tink-worker/tink-worker-linux-amd64
	docker build -t tink-worker cmd/tink-worker/
virtual-worker-image: cmd/virtual-worker/virtual-worker-linux-amd64
	docker build -t virtual-worker cmd/virtual-worker/

.PHONY: run-stack
run-stack:
	docker-compose up --build

ifeq ($(origin GOBIN), undefined)
GOBIN := ${PWD}/bin
export GOBIN
PATH := ${GOBIN}:${PATH}
export PATH
endif

toolsBins := $(addprefix bin/,$(notdir $(shell awk -F'"' '/^\s*_/ {print $$2}' tools.go)))

# installs cli tools defined in tools.go
$(toolsBins): go.mod go.sum tools.go
$(toolsBins): CMD=$(shell awk -F'"' '/$(@F)"/ {print $$2}' tools.go)
$(toolsBins):
	go install $(CMD)

.PHONY: protomocks
protomocks: bin/moq
	go generate ./protos/...
	gofumpt -s -w ./protos/*/mock.go

.PHONY: check-protomocks
check-protomocks:
	@git diff --no-ext-diff --quiet --exit-code -- protos/*/mock.go || (
	  echo "Mock files need to be regenerated!";
	  git diff --no-ext-diff --exit-code --stat -- protos/*/mock.go
	)

.PHONY: pbfiles
pbfiles: buf.gen.yaml buf.lock $(shell git ls-files 'protos/*/*.proto') $(toolsBins)
	buf generate
	gofumpt -w protos/*/*.pb.*

.PHONY: check-pbfiles
check-pbfiles: pbfiles
	@git diff --no-ext-diff --quiet --exit-code -- protos/*/*.pb.* || (
	  echo "Protobuf files need to be regenerated!";
	  git diff --no-ext-diff --exit-code --stat -- protos/*/*.pb.*
	)

e2etest-setup: $(toolsBins)
	setup-envtest use
