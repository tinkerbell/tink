all: cli server worker ## Build all binaries for host OS and CPU

-include rules.mk

crosscompile: $(crossbinaries) ## Build all binaries for Linux and all supported CPU arches
images: tink-cli-image tink-server-image tink-worker-image ## Build all docker images
run: crosscompile run-stack ## Builds and runs the Tink stack (tink, db, cli) via docker-compose

test: ## Run tests
	go clean -testcache
	go test ./... -v

verify: ## Run lint like checkers
	goimports -d .
	golint ./...

help: ## Print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120


GOLINT_VERSION ?= v1.41.1
HADOLINT_VERSION ?= v2.6.1
SHELLCHECK_VERSION ?= v0.7.2
OS := $(shell uname)
LOWER_OS  = $(shell echo $OS | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

lint: out/linters/shellcheck-$(SHELLCHECK_VERSION)/shellcheck out/linters/hadolint-$(HADOLINT_VERSION) out/linters/golangci-lint-$(GOLINT_VERSION)
	./out/linters/golangci-lint-$(GOLINT_VERSION) run
	out/linters/shellcheck-$(SHELLCHECK_VERSION)/shellcheck $(shell find . -name "*.sh")
	out/linters/hadolint-$(HADOLINT_VERSION) $(shell find . -name "*Dockerfile")

out/linters/shellcheck-$(SHELLCHECK_VERSION)/shellcheck:
	mkdir -p out/linters
	curl -sfL https://github.com/koalaman/shellcheck/releases/download/v0.7.2/shellcheck-$(SHELLCHECK_VERSION).$(OS).$(ARCH).tar.xz | tar -C out/linters -zxvf -

out/linters/hadolint-$(HADOLINT_VERSION):
	mkdir -p out/linters
	curl -sfL https://github.com/hadolint/hadolint/releases/download/v2.6.1/hadolint-$(OS)-$(ARCH) > out/linters/hadolint-$(HADOLINT_VERSION)
	chmod u+x out/linters/hadolint-$(HADOLINT_VERSION)

out/linters/golangci-lint-$(GOLINT_VERSION):
	mkdir -p out/linters
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b out/linters $(GOLINT_VERSION)
	mv out/linters/golangci-lint out/linters/golangci-lint-$(GOLINT_VERSION)
