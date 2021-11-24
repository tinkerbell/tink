GO_INSTALL = ./scripts/go_install.sh
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)

CONTROLLER_GEN_VER := v0.7.0
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)

ENVSUBST_BIN := envsubst
ENVSUBST := $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)-drone

# --------------------------------------
# Tooling Binaries
# --------------------------------------

$(CONTROLLER_GEN): # Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)

.PHONY: generate
generate: # Generate code, manifests etc.
	$(MAKE) generate-go
	$(MAKE) generate-manifests

.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) # Generate Go code.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./pkg/apis/..."

.PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) # Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./pkg/apis/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
