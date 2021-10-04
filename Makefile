all: cli server worker ## Build all binaries for host OS and CPU

-include rules.mk

crosscompile: $(crossbinaries) ## Build all binaries for Linux and all supported CPU arches
images: tink-cli-image tink-server-image tink-worker-image ## Build all docker images
run: crosscompile run-stack ## Builds and runs the Tink stack (tink, db, cli) via docker-compose

test: ## Run tests
	go clean -testcache
	go test ./... -v

verify: lint

help: ## Print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120

# BEGIN: lint-install --dockerfile=warn .
# http://github.com/tinkerbell/lint-install

GOLINT_VERSION ?= v1.42.0
HADOLINT_VERSION ?= v2.7.0
SHELLCHECK_VERSION ?= v0.7.2
LINT_OS := $(shell uname)
LINT_ARCH := $(shell uname -m)
LINT_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# shellcheck and hadolint lack arm64 native binaries: rely on x86-64 emulation
ifeq ($(LINT_OS),Darwin)
	ifeq ($(LINT_ARCH),arm64)
		LINT_ARCH=x86_64
	endif
endif

LINT_LOWER_OS  = $(shell echo $(LINT_OS) | tr '[:upper:]' '[:lower:]')
GOLINT_CONFIG:=$(LINT_ROOT)/.golangci.yml

lint: out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)/shellcheck out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH) out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH)
	out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH) run
	out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH) --no-fail $(shell find . -name "*Dockerfile")
	out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)/shellcheck $(shell find . -name "*.sh")

fix: out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)/shellcheck out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH)
	out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH) run --fix
	out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)/shellcheck $(shell find . -name "*.sh") -f diff | git apply -p2 -

out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)/shellcheck:
	mkdir -p out/linters
	curl -sSfL https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/shellcheck-$(SHELLCHECK_VERSION).$(LINT_LOWER_OS).$(LINT_ARCH).tar.xz | tar -C out/linters -xJf -
	mv out/linters/shellcheck-$(SHELLCHECK_VERSION) out/linters/shellcheck-$(SHELLCHECK_VERSION)-$(LINT_ARCH)

out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH):
	mkdir -p out/linters
	curl -sfL https://github.com/hadolint/hadolint/releases/download/v2.6.1/hadolint-$(LINT_OS)-$(LINT_ARCH) > out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH)
	chmod u+x out/linters/hadolint-$(HADOLINT_VERSION)-$(LINT_ARCH)

out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH):
	mkdir -p out/linters
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b out/linters $(GOLINT_VERSION)
	mv out/linters/golangci-lint out/linters/golangci-lint-$(GOLINT_VERSION)-$(LINT_ARCH)

# END: lint-install --dockerfile=warn .
