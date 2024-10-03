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

include Tools.mk

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
BINARIES := tink-server tink-agent tink-worker tink-controller tink-controller-v1alpha2 virtual-worker

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
e2e-test: $(SETUP_ENVTEST)
	$(SETUP_ENVTEST) use
	source <($(SETUP_ENVTEST) use -p env) && $(GO) test -v ./internal/e2e/... -tags=e2e

mocks: $(MOQ)
	$(MOQ) -fmt goimpots -rm -out ./internal/proto/workflow/v2/mock.go ./internal/proto/workflow/v2 WorkflowServiceClient WorkflowService_GetWorkflowsClient
	$(MOQ) -fmt goimports -rm -out ./internal/agent/transport/mock.go ./internal/agent/transport WorkflowHandler
	$(MOQ) -fmt goimports -rm -out ./internal/agent/mock.go ./internal/agent Transport ContainerRuntime
	$(MOQ) -fmt goimports -rm -out ./internal/agent/event/mock.go ./internal/agent/event Recorder

.PHONY: generate-proto
generate-proto: buf.gen.yaml buf.lock $(shell git ls-files '**/*.proto') $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(GOFUMPT)
	$(BUF) mod update
	$(BUF) generate
	$(GOFUMPT) -w internal/proto/*.pb.*
	$(GOFUMPT) -w internal/proto/workflow/v2/*.pb.*

.PHONY: generate
generate: ## Generate code, manifests etc.
generate: generate-proto generate-go generate-manifests

.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) $(GOFUMPT)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./api/..."
	$(GOFUMPT) -w ./api

.PHONY: generate-manifests
generate-manifests: ## Generate manifests e.g. CRD, RBAC etc.
generate-manifests: generate-crds generate-rbac

.PHONY: generate-crds
generate-crds: $(CONTROLLER_GEN) $(YAMLFMT)
	$(CONTROLLER_GEN) \
		paths=./api/... \
		crd:crdVersions=v1 \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
	$(YAMLFMT) ./config/crd/bases/* ./config/webhook/*

.PHONY: generate-rbac
generate-rbac: generate-controller-rbac generate-server-rbac $(CONTROLLER_GEN) $(YAMLFMT)

.PHONY: generate-controller-rbac
generate-controller-rbac:
	$(CONTROLLER_GEN) \
		paths=./internal/deprecated/workflow/... \
		output:rbac:dir=./config/manager-rbac/ \
		rbac:roleName=manager-role
	$(YAMLFMT) ./config/rbac/*

.PHONY: generate-server-rbac
generate-server-rbac: $(CONTROLLER_GEN) $(YAMLFMT)
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

out/release/tink.yaml: generate-manifests out/release/default/kustomization.yaml $(KUSTOMIZE) $(YAMLFMT)
	(
		cd out/release/default && \
		$(KUSTOMIZE) edit set image server=$(TINK_SERVER_IMAGE):$(TINK_CONTROLLER_TAG) controller=$(TINK_CONTROLLER_IMAGE):$(TINK_CONTROLLER_TAG) && \
		$(KUSTOMIZE) edit set namespace $(NAMESPACE) \
	)
	$(KUSTOMIZE) build out/release/default -o $@
	$(YAMLFMT) $@

.PHONY: release-manifests
release-manifests: ## Builds the manifests to publish with a release.
release-manifests: out/release/tink.yaml

.PHONY: check-generated
check-generated: ## Check if generated files are up to date.
check-generated: check-proto

.PHONY: check-proto
check-proto: generate-proto
	@git diff --no-ext-diff --quiet --exit-code -- **/*.pb.* || (
	  echo "Protobuf files need to be regenerated!";
	  git diff --no-ext-diff --exit-code -- **/*.pb.*
	)

.PHONY: verify
verify: ## Verify code style, is lint free, freshness ...
verify: lint check-generated $(GOFUMPT)
	$(GOFUMPT) -d .

.PHONY: ci-checks
ci-checks: ## Run ci-checks.sh script
	@if type nix-shell 2>&1 > /dev/null; then \
		./ci-checks.sh; \
	else \
		docker run -it --rm -v $${PWD}:/code -w /code nixos/nix nix-shell --run 'make ci-checks'; \
	fi

.PHONY: lint
lint: ## Lint code.
lint: shellcheck hadolint golangci-lint yamllint

.PHONY: shellcheck
shellcheck: $(SHELLCHECK)
	$(SHELLCHECK) $(shell find . -name "*.sh")

.PHONY: hadolint
hadolint: $(HADOLINT)
	$(HADOLINT) --no-fail $(shell find . -name "*Dockerfile")

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

.PHONY: yamllint
yamllint: $(YAMLLINT_BIN)
	$(YAMLLINT) .
