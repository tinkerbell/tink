# Define the directory tools are installed to.
TOOLS_DIR := $(PWD)/out/tools

# Some tools rely on other tools being on the PATH (protoc).
PATH := $(PATH):$(TOOLS_DIR)

# Define a variable for Go so we can easily change the version we use.
GO ?= go

LINT_ARCH := $(shell uname -m)
LINT_OS := $(shell uname)
LINT_OS_LOWER := $(shell echo $(LINT_OS) | tr '[:upper:]' '[:lower:]')

BUF_VER := v1.11
BUF     := $(TOOLS_DIR)/buf-$(BUF_VER)

GOFUMPT_VER := v0.4
GOFUMPT     := $(TOOLS_DIR)/gofumpt-$(GOFUMPT_VER)

PROTOC_GEN_GO_GRPC_VER := v1.2
PROTOC_GEN_GO_GRPC     := $(TOOLS_DIR)/protoc-gen-go-grpc

PROTOC_GEN_GO_VER := v1.28
PROTOC_GEN_GO     := $(TOOLS_DIR)/protoc-gen-go

CONTROLLER_GEN_VER := v0.16.3
CONTROLLER_GEN     := $(TOOLS_DIR)/controller-gen-$(CONTROLLER_GEN_VER)

KUSTOMIZE_VER := v4.5
KUSTOMIZE     := $(TOOLS_DIR)/kustomize-$(KUSTOMIZE_VER)

SETUP_ENVTEST_VER := v0.0.0-20220304125252-9ee63fc65a97
SETUP_ENVTEST     := $(TOOLS_DIR)/setup-envtest-$(SETUP_ENVTEST_VER)

GOLANGCI_LINT_VER := v1.52
GOLANGCI_LINT     := $(TOOLS_DIR)/golangci-lint-$(GOLANGCI_LINT_VER)

YAMLFMT_VER := v0.6
YAMLFMT     := $(TOOLS_DIR)/yamlfmt-$(YAMLFMT_VER)

MOQ_VER := v0.3
MOQ     := $(TOOLS_DIR)/moq-$(MOQ_VER)

SHELLCHECK_VER := v0.8.0
SHELLCHECK     := $(TOOLS_DIR)/shellcheck-$(SHELLCHECK_VER)

HADOLINT_VER := v2.12.1-beta
HADOLINT     := $(TOOLS_DIR)/hadolint-$(HADOLINT_VER)

YAMLLINT_VER  := 1.26.3
YAMLLINT_ROOT := $(TOOLS_DIR)/yamllint-$(YAMLLINT_VER)
YAMLLINT_BIN  := $(YAMLLINT_ROOT)/dist/bin/yamllint
YAMLLINT      := PYTHONPATH=$(YAMLLINT_ROOT)/dist $(YAMLLINT_ROOT)/dist/bin/yamllint

tools: ## Install all tool depedencies.
tools: $(BUF) $(GOFUMPT) $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GO) $(CONTROLLER_GEN) $(KUSTOMIZE)
tools: $(SETUP_ENVTEST) $(GOLANGCI_LINT) $(YAMLFMT) $(MOQ) $(SHELLCHECK) $(HADOLINT) $(YAMLLINT_BIN)

$(BUF):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing buf at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install github.com/bufbuild/buf/cmd/buf@$(BUF_VER)
	@mv $(TOOLS_DIR)/buf $@

$(GOFUMPT):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing gofumpt at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install mvdan.cc/gofumpt@$(GOFUMPT_VER)
	@mv $(TOOLS_DIR)/gofumpt $@

$(PROTOC_GEN_GO_GRPC):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing protoc-gen-go-grpc at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VER)

$(PROTOC_GEN_GO):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing protoc-gen-go at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VER)

$(CONTROLLER_GEN):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing controller-gen at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VER)
	@mv $(TOOLS_DIR)/controller-gen $@

$(KUSTOMIZE):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing kustomize at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install sigs.k8s.io/kustomize/kustomize/v4@$(KUSTOMIZE_VER)
	@mv $(TOOLS_DIR)/kustomize $@

$(SETUP_ENVTEST):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing setup-envtest at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(SETUP_ENVTEST_VER)
	@mv $(TOOLS_DIR)/setup-envtest $@

$(GOLANGCI_LINT):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing golangci-lint at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VER)
	@mv $(TOOLS_DIR)/golangci-lint $@

$(YAMLFMT):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing yamlfmt at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install github.com/google/yamlfmt/cmd/yamlfmt@$(YAMLFMT_VER)
	@mv $(TOOLS_DIR)/yamlfmt $@

$(MOQ):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing moq at $@"
	@GOBIN=$(TOOLS_DIR) $(GO) install github.com/matryer/moq@$(MOQ_VER)
	@mv $(TOOLS_DIR)/moq $@

$(SHELLCHECK):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing shellcheck at $@"
	@mkdir -p _tmp
	@curl -sSfL https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VER)/shellcheck-$(SHELLCHECK_VER).$(LINT_OS_LOWER).$(LINT_ARCH).tar.xz | \
		tar -C _tmp -xJ --strip-components 1 -f -
	@mv _tmp/shellcheck $@
	@rm -rf _tmp
	@chmod u+x $@

$(HADOLINT):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing hadolint at $@"
	@curl -sSfL -o $@ https://github.com/hadolint/hadolint/releases/download/$(HADOLINT_VER)/hadolint-$(LINT_OS)-$(LINT_ARCH)
	@chmod u+x $@

$(YAMLLINT_BIN):
	@mkdir -p $(TOOLS_DIR)
	@echo "Installing yamllint at $@"
	@curl -sSfL https://github.com/adrienverge/yamllint/archive/refs/tags/v$(YAMLLINT_VER).tar.gz | \
		tar -C $(TOOLS_DIR) -zxf -
	@cd $(YAMLLINT_ROOT) && pip3 install --target dist . > /dev/null || pip install --target dist .
