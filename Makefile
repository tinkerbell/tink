# Only use the recipes defined in these makefiles
MAKEFLAGS += --no-builtin-rules

PATH  := $(PATH):$(PWD)/bin

# Use bash instead of plain sh and treat the shell as one shell script invocation.
SHELL 		:= bash
.SHELLFLAGS := -o pipefail -euc
.ONESHELL:

# Second expansion is used by the image targets to depend on their respective binaries. It is
# necessary because automatic variables are not set on first expansion.
# See https://www.gnu.org/software/make/manual/html_node/Secondary-Expansion.html.
.SECONDEXPANSION:

# Configure Go environment variables for use in the Makefile.
GOARCH 	?= $(shell go env GOARCH)
GOOS 	?= $(shell go env GOOS)
GOPROXY ?= $(shell go env GOPROXY)

# Runnable tools
GO 				?= go
BUF 			:= $(GO) run github.com/bufbuild/buf/cmd/buf@v1.11
CONTROLLER_GEN 	:= $(GO) run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11
GOFUMPT 		:= $(GO) run mvdan.cc/gofumpt@v0.4
KUSTOMIZE 		:= $(GO) run sigs.k8s.io/kustomize/kustomize/v4@v4.5
SETUP_ENVTEST   := $(GO) run sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20220304125252-9ee63fc65a97
GOLANGCI_LINT	:= $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52
YAMLFMT			:= $(GO) run github.com/google/yamlfmt/cmd/yamlfmt@v0.6

# Installed tools
PROTOC_GEN_GO_GRPC 	:= google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
PROTOC_GEN_GO 		:= google.golang.org/protobuf/cmd/protoc-gen-go@v1.28

.PHONY: help
help: ## Print this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[%\/0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo
	@echo Individual binaries can be built with their name. For example, \`make tink-server\`.
	@echo
	@echo Individual images can be built with their name appended with -image. For example, 
	@echo \`make tink-server-image\`.

# Version defines the string injected into binaries that indicates the version of the build.
# Eventually this will be superfluous as we transition to buildinfo provided by the Go standard
# library.
# Override by setting VERSION when invoking make. For example, `make VERSION=1.2.3`.
VERSION ?= $(shell git rev-parse --short HEAD)

# Define all the binaries we build for this project that get packaged into containers.
BINARIES := tink-server tink-worker tink-controller virtual-worker

.PHONY: build
build: $(BINARIES) ## Build all tink binaries. Cross build by setting GOOS and GOARCH.

# Create targets for all the binaries we build. They can be individually invoked with `make <binary>`.
# For example, `make tink-server`. Callers can cross build by defining the GOOS and GOARCH 
# variables. For example, `GOOS=linux GOARCH=arm64 make tink-server`.
# See https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html.
.PHONY: $(BINARIES)
$(BINARIES):
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LDFLAGS) -o ./bin/$@-$(GOOS)-$(GOARCH) ./cmd/$@

# IMAGE_ARGS is resolved when its used in the `%-image` targets. Consequently, the $* automatic 
# variable isn't evaluated until the target is called.  
# See https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html.
IMAGE_ARGS ?= -t $*

.PHONY: images
images: $(addsuffix -image,$(BINARIES)) ## Build all tink container images. All images will be for Linux but target the host arch.

# Image building targets are used for local builds only. The CI builds images using
# Github actions as they facilitate build and push functions, and have better support for
# release tagging etc.
#
# We only build Linux images so we need to force binaries to be built for Linux. Exporting the
# GOOS variable ensures the recipe's binary dependency is built for Linux.
#
# The $$* leverages .SECONDEXPANSION to specify the matched part of the target name as a 
# dependency. In doing so, we ensure the binary is built so it can be copied into the image. For 
# example, `make tink-server-image` will depend on `tink-server`.
# See https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html.
# See https://www.gnu.org/software/make/manual/html_node/Secondary-Expansion.html.
%-image: export GOOS=linux
%-image: $$*
	DOCKER_BUILDKIT=1 docker build $(IMAGE_ARGS) -f cmd/$*/Dockerfile .

.PHONY: test
test: ## Run tests
	$(GO) test -coverprofile=coverage.txt ./...

.PHONY: e2e-test
e2e-test: ## Run e2e tests
	$(SETUP_ENVTEST) use
	source <($(SETUP_ENVTEST) use -p env) && $(GO) test -v ./internal/e2e/... -tags=e2e

.PHONY: generate-proto
generate-proto: buf.gen.yaml buf.lock $(shell git ls-files '**/*.proto') _protoc
	$(BUF) mod update
	$(BUF) generate
	$(GOFUMPT) -w internal/proto/*.pb.*

.PHONY: generate
generate: generate-proto generate-go generate-manifests ## Generate code, manifests etc.

.PHONY: generate-go
generate-go:
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./api/..."
	$(GOFUMPT) -w ./api

.PHONY: generate-manifests
generate-manifests: generate-crds generate-rbacs generate-server-rbacs ## Generate manifests e.g. CRD, RBAC etc.

.PHONY: generate-crds
generate-crds:
	$(CONTROLLER_GEN) \
		paths=./api/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
	$(YAMLFMT) ./config/crd/bases/* ./config/webhook/*

.PHONY: generate-rbacs
generate-rbacs:
	$(CONTROLLER_GEN) \
		paths=./internal/controller/... \
		output:rbac:dir=./config/rbac/ \
		rbac:roleName=manager-role
	$(YAMLFMT) ./config/rbac/*

.PHONY: generate-server-rbacs
generate-server-rbacs:
	$(CONTROLLER_GEN) \
		paths=./internal/server/... \
		output:rbac:dir=./config/server-rbac \
		rbac:roleName=server-role
	$(YAMLFMT) ./config/server-rbac/*

TINK_SERVER_IMAGE 	?= quay.io/tinkerbell/tink-server
TINK_SERVER_TAG 	?= latest

TINK_CONTROLLER_IMAGE 	?= quay.io/tinkerbell/tink-controller
TINK_CONTROLLER_TAG 	?= latest

NAMESPACE ?= tink-system

out/release/default/kustomization.yaml: config/default/kustomization.yaml
	rm -rf out/
	mkdir -p out/
	cp -a config/ out/release/

out/release/tink.yaml: generate-manifests out/release/default/kustomization.yaml
	(
		cd out/release/default && \
		$(KUSTOMIZE) edit set image server=$(TINK_SERVER_IMAGE):$(TINK_CONTROLLER_TAG) controller=$(TINK_CONTROLLER_IMAGE):$(TINK_CONTROLLER_TAG) && \
		$(KUSTOMIZE) edit set namespace $(NAMESPACE) \
	)
	$(KUSTOMIZE) build out/release/default -o $@
	prettier --write $@

.PHONY: release-manifests
release-manifests: out/release/tink.yaml ## Builds the manifests to publish with a release.

.PHONY: check-generated
check-generated: check-proto ## Check if generated files are up to date.

.PHONY: check-proto
check-proto: generate-proto
	@git diff --no-ext-diff --quiet --exit-code -- **/*.pb.* || (
	  echo "Protobuf files need to be regenerated!";
	  git diff --no-ext-diff --exit-code -- **/*.pb.*
	)

.PHONY: verify
verify: lint check-generated ## Verify code style, is lint free, freshness ...
	$(GOFUMPT) -d .

.PHONY: ci-checks
ci-checks: ## Run ci-checks.sh script
	@if type nix-shell 2>&1 > /dev/null; then \
		./ci-checks.sh; \
	else \
		docker run -it --rm -v $${PWD}:/code -w /code nixos/nix nix-shell --run 'make ci-checks'; \
	fi

.PHONY: lint
lint: shellcheck hadolint golangci-lint yamllint ## Lint code

LINT_ARCH := $(shell uname -m)
LINT_OS := $(shell uname)
LINT_OS_LOWER := $(shell echo $(LINT_OS) | tr '[:upper:]' '[:lower:]')

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

.PHONY: shellcheck
shellcheck: $(SHELLCHECK_BIN)
	$(SHELLCHECK_BIN) $(shell find . -name "*.sh")

HADOLINT_VERSION ?= v2.12.1-beta
HADOLINT_BIN := out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH)
$(HADOLINT_BIN):
	mkdir -p out/linters
	curl -sSfL -o $@.dl https://github.com/hadolint/hadolint/releases/download/$(HADOLINT_VERSION)/hadolint-$(LINT_OS)-$(LINT_ARCH) \
		|| echo "Unable to fetch hadolint for $(LINT_OS)/$(LINT_ARCH), falling back to local install"
	test -f $@.dl && mv $(HADOLINT_BIN).dl $@ || printf "#!/usr/bin/env hadolint\n" > $@
	chmod u+x $@

.PHONY: hadolint
hadolint: $(HADOLINT_BIN)
	$(HADOLINT_BIN) --no-fail $(shell find . -name "*Dockerfile")

.PHONY: golangci-lint
golangci-lint:
	$(GOLANGCI_LINT) run

YAMLLINT_VERSION ?= 1.26.3
YAMLLINT_ROOT := out/linters/yamllint-$(YAMLLINT_VERSION)
YAMLLINT_BIN := $(YAMLLINT_ROOT)/dist/bin/yamllint
$(YAMLLINT_BIN):
	mkdir -p out/linters
	rm -rf out/linters/yamllint-*
	curl -sSfL https://github.com/adrienverge/yamllint/archive/refs/tags/v$(YAMLLINT_VERSION).tar.gz | tar -C out/linters -zxf -
	cd $(YAMLLINT_ROOT) && pip3 install --target dist . || pip install --target dist .

.PHONY: yamllint
yamllint: $(YAMLLINT_BIN)
	PYTHONPATH=$(YAMLLINT_ROOT)/dist $(YAMLLINT_ROOT)/dist/bin/yamllint .

.PHONY: _protoc ## Install all required tools for use with this Makefile.
_protoc:
	$(GO) install $(PROTOC_GEN_GO)
	$(GO) install $(PROTOC_GEN_GO_GRPC)