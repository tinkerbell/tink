help: ## Print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120

all: cli server worker ## Build all binaries for host OS and CPU

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

## Tools

GO 				:= go
BUF 			:= $(GO) run github.com/bufbuild/buf/cmd/buf@v1.11.0
CONTROLLER_GEN 	:= $(GO) run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10
GOFUMPT 		:= $(GO) run mvdan.cc/gofumpt@v0.4
KUSTOMIZE 		:= $(GO) run sigs.k8s.io/kustomize/kustomize/v4@latestv4.5

####

binaries := cmd/tink-controller/tink-controller cmd/tink-server/tink-server cmd/tink-worker/tink-worker cmd/virtual-worker/virtual-worker
version := $(shell git rev-parse --short HEAD)
tag := $(shell git tag --points-at HEAD | head -n 1)
ifneq (,$(tag))
version := $(tag)-$(version)
endif
LDFLAGS := -ldflags "-X main.version=$(version)"
export CGO_ENABLED := 0

.PHONY: server cli worker virtual-worker test $(binaries)
controller: cmd/tink-controller/tink-controller
server: cmd/tink-server/tink-server
worker : cmd/tink-worker/tink-worker
virtual-worker : cmd/virtual-worker/virtual-worker

crossbinaries := $(addsuffix -linux-,$(binaries))
crossbinaries := $(crossbinaries:=amd64) $(crossbinaries:=arm64)

.PHONY: crosscompile $(crossbinaries)
%-amd64: FLAGS=GOOS=linux GOARCH=amd64
%-arm64: FLAGS=GOOS=linux GOARCH=arm64
$(binaries) $(crossbinaries):
	$(FLAGS) $(GO) build $(LDFLAGS) -o $@ ./$(@D)

.PHONY: tink-controller-image tink-server-image tink-worker-image virtual-worker-image
tink-controller-image: cmd/tink-controller/tink-controller-linux-amd64
	docker build -t tink-controller cmd/tink-controller/
tink-server-image: cmd/tink-server/tink-server-linux-amd64
	docker build -t tink-server cmd/tink-server/
tink-worker-image: cmd/tink-worker/tink-worker-linux-amd64
	docker build -t tink-worker cmd/tink-worker/
virtual-worker-image: cmd/virtual-worker/virtual-worker-linux-amd64
	docker build -t virtual-worker cmd/virtual-worker/

ifeq ($(origin GOBIN), undefined)
GOBIN := ${PWD}/bin
export GOBIN
PATH := ${GOBIN}:${PATH}
export PATH
endif

# These need changing to a `go run` command and placed in the Tools section of the Makefile
.PHONY: tools ## Install all required tools for use with this Makefile.
tools:
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	$(GO) install sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20220304125252-9ee63fc65a97
	$(GO) install sigs.k8s.io/kustomize/kustomize/v4@v4.5.4

.PHONY: pbfiles
pbfiles: buf.gen.yaml buf.lock $(shell git ls-files 'protos/*/*.proto') tools
	$(BUF) mod update
	$(BUF) generate
	$(GOFUMPT) -w protos/*/*.pb.*

.PHONY: check-pbfiles
check-pbfiles: pbfiles
	@git diff --no-ext-diff --quiet --exit-code -- protos/*/*.pb.* || (
	  echo "Protobuf files need to be regenerated!";
	  git diff --no-ext-diff --exit-code -- protos/*/*.pb.*
	)

e2etest-setup: tools
	setup-envtest use

## --------------------------------------
## Generate
## --------------------------------------

.PHONY: generate
generate: generate-go generate-manifests # Generate code, manifests etc.

.PHONY: generate-go
generate-go: # Generate Go code.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./pkg/apis/..."
	$(GOFUMPT) -w ./pkg/apis

.PHONY: generate-manifests
generate-manifests: generate-crds generate-rbacs generate-server-rbacs # Generate manifests e.g. CRD, RBAC etc.

.PHONY: generate-crds
generate-crds:
	$(CONTROLLER_GEN) \
		paths=./pkg/apis/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook

.PHONY: generate-rbacs
generate-rbacs: 
	$(CONTROLLER_GEN) \
		paths=./pkg/controllers/... \
		output:rbac:dir=./config/rbac/ \
		rbac:roleName=manager-role

.PHONY: generate-server-rbacs
generate-server-rbacs: 
	$(CONTROLLER_GEN) \
		paths=./server/... \
		output:rbac:dir=./config/server-rbac \
		rbac:roleName=server-role

TINK_SERVER_IMAGE 	?= quay.io/tinkerbell/tink-server
TINK_SERVER_TAG 	?= latest

TINK_CONTROLLER_IMAGE 	?= quay.io/tinkerbell/tink-controller
TINK_CONTROLLER_TAG 	?= latest

NAMESPACE ?= tink-system

out/release/default/kustomization.yaml: config/default/kustomization.yaml
	rm -rf out/
	mkdir -p out/
	cp -a config/ out/release/

out/release/tink.yaml: bin/kustomize generate-manifests out/release/default/kustomization.yaml
	(
		cd out/release/default && \
		$(KUSTOMIZE) edit set image server=$(TINK_SERVER_IMAGE):$(TINK_CONTROLLER_TAG) controller=$(TINK_CONTROLLER_IMAGE):$(TINK_CONTROLLER_TAG) && \
		$(KUSTOMIZE) edit set namespace $(NAMESPACE) \
	)
	$(KUSTOMIZE) build out/release/default -o $@
	prettier --write $@

release-manifests: out/release/tink.yaml ## Builds the manifests to publish with a release.

crosscompile: $(crossbinaries) ## Build all binaries for Linux and all supported CPU arches
images:  tink-server-image tink-worker-image virtual-worker-image ## Build all docker images
run: crosscompile run-stack ## Builds and runs the Tink stack (tink, db, cli) via docker-compose

test: e2etest-setup ## Run tests
	source <(setup-envtest use -p env) && $(GO) test -coverprofile=coverage.txt ./...

verify: lint check-generated ## Verify code style, is lint free, freshness ...
	$(GOFUMPT) -s -d .

generated: pbfiles generate-manifests ## Generate dynamically created files
check-generated: check-pbfiles ## Check if generated files are up to date

.PHONY: all check-generated crosscompile generated help images run test tools verify

# BEGIN: lint-install --dockerfile=warn -makefile=lint.mk .
# http://github.com/tinkerbell/lint-install

.PHONY: lint
lint: _lint

LINT_ARCH := $(shell uname -m)
LINT_OS := $(shell uname)
LINT_OS_LOWER := $(shell echo $(LINT_OS) | tr '[:upper:]' '[:lower:]')
LINT_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# shellcheck and hadolint lack arm64 native binaries: rely on x86-64 emulation
ifeq ($(LINT_OS),Darwin)
	ifeq ($(LINT_ARCH),arm64)
		LINT_ARCH=x86_64
	endif
endif

LINTERS :=
FIXERS :=

SHELLCHECK_VERSION ?= v0.8.0
SHELLCHECK_BIN := out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)
$(SHELLCHECK_BIN):
	mkdir -p out/linters
	curl -sSfL -o $@.tar.xz https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/shellcheck-$(SHELLCHECK_VERSION).$(LINT_OS_LOWER).$(LINT_ARCH).tar.xz \
		|| echo "Unable to fetch shellcheck for $(LINT_OS)/$(LINT_ARCH): falling back to locally install"
	test -f $@.tar.xz \
		&& tar -C out/linters -xJf $@.tar.xz \
		&& mv out/linters/shellcheck-$(SHELLCHECK_VERSION)/shellcheck $@ \
		|| printf "#!/usr/bin/env shellcheck\n" > $@
	chmod u+x $@

LINTERS += shellcheck-lint
shellcheck-lint: $(SHELLCHECK_BIN)
	$(SHELLCHECK_BIN) $(shell find . -name "*.sh")

FIXERS += shellcheck-fix
shellcheck-fix: $(SHELLCHECK_BIN)
	$(SHELLCHECK_BIN) $(shell find . -name "*.sh") -f diff | { read -t 1 line || exit 0; { echo "$$line" && cat; } | git apply -p2; }

HADOLINT_VERSION ?= v2.8.0
HADOLINT_BIN := out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH)
$(HADOLINT_BIN):
	mkdir -p out/linters
	curl -sSfL -o $@.dl https://github.com/hadolint/hadolint/releases/download/$(HADOLINT_VERSION)/hadolint-$(LINT_OS)-$(LINT_ARCH) \
		|| echo "Unable to fetch hadolint for $(LINT_OS)/$(LINT_ARCH), falling back to local install"
	test -f $@.dl && mv $(HADOLINT_BIN).dl $@ || printf "#!/usr/bin/env hadolint\n" > $@
	chmod u+x $@

LINTERS += hadolint-lint
hadolint-lint: $(HADOLINT_BIN)
	$(HADOLINT_BIN) --no-fail $(shell find . -name "*Dockerfile")

GOLANGCI_LINT_CONFIG := $(LINT_ROOT)/.golangci.yml
GOLANGCI_LINT_VERSION ?= v1.49.0
GOLANGCI_LINT_BIN := out/linters/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(LINT_ARCH)
$(GOLANGCI_LINT_BIN):
	mkdir -p out/linters
	rm -rf out/linters/golangci-lint-*
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b out/linters $(GOLANGCI_LINT_VERSION)
	mv out/linters/golangci-lint $@

LINTERS += golangci-lint-lint
golangci-lint-lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run

FIXERS += golangci-lint-fix
golangci-lint-fix: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run --fix

YAMLLINT_VERSION ?= 1.26.3
YAMLLINT_ROOT := out/linters/yamllint-$(YAMLLINT_VERSION)
YAMLLINT_BIN := $(YAMLLINT_ROOT)/dist/bin/yamllint
$(YAMLLINT_BIN):
	mkdir -p out/linters
	rm -rf out/linters/yamllint-*
	curl -sSfL https://github.com/adrienverge/yamllint/archive/refs/tags/v$(YAMLLINT_VERSION).tar.gz | tar -C out/linters -zxf -
	cd $(YAMLLINT_ROOT) && pip3 install --target dist . || pip install --target dist .

LINTERS += yamllint-lint
yamllint-lint: $(YAMLLINT_BIN)
	PYTHONPATH=$(YAMLLINT_ROOT)/dist $(YAMLLINT_ROOT)/dist/bin/yamllint .

.PHONY: _lint $(LINTERS)
_lint: $(LINTERS)

.PHONY: fix $(FIXERS)
fix: $(FIXERS)

# END: lint-install --dockerfile=warn -makefile=lint.mk .
