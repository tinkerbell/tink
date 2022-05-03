## --------------------------------------
## Generate
## --------------------------------------

PHONY: generate
generate: generate-go generate-manifests # Generate code, manifests etc.

.PHONY: generate-go
generate-go: bin/controller-gen bin/gofumpt # Generate Go code.
	controller-gen object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./pkg/apis/..."
	gofumpt -w -s ./pkg/apis

.PHONY: generate-manifests
generate-manifests: generate-crds generate-rbacs generate-server-rbacs # Generate manifests e.g. CRD, RBAC etc.

.PHONY: generate-crds
generate-crds: bin/controller-gen
	controller-gen \
		paths=./pkg/apis/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
	prettier --write ./config/crd/bases

.PHONY: generate-rbacs
generate-rbacs: bin/controller-gen
	controller-gen \
		paths=./pkg/controllers/... \
		output:rbac:dir=./config/rbac/ \
		rbac:roleName=manager-role
	prettier --write ./config/rbac

.PHONY: generate-server-rbacs
generate-server-rbacs: bin/controller-gen
	controller-gen \
		paths=./server/... \
		output:rbac:dir=./config/server-rbac \
		rbac:roleName=server-role
	prettier --write ./config/server-rbac/

RELEASE_DIR ?= out/release

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

REGISTRY ?= quay.io/tinkerbell
TINK_SERVER_IMAGE_NAME ?= tink-server
TINK_CONTROLLER_IMAGE_NAME ?= tink-controller
TINK_SERVER_IMAGE_TAG ?= latest
TINK_CONTROLLER_IMAGE_TAG ?= latest

.PHONY: set-manager-manifest-image
set-manager-manifest-image:
	$(info Updating kustomize image patch file for tink-controller)
	sed -i'' -e 's@image: .*@image: '"${MANIFEST_IMG}:$(MANIFEST_TAG)"'@' ./config/default/manager_image_patch.yaml

.PHONY: set-server-manifest-image
set-server-manifest-image:
	$(info Updating kustomize image patch file for tink-server)
	sed -i'' -e 's@image: .*@image: '"${MANIFEST_IMG}:$(MANIFEST_TAG)"'@' ./config/default/server_image_patch.yaml

.PHONY: release
release: clean-release
	$(MAKE) set-manager-manifest-image MANIFEST_IMG=$(REGISTRY)/$(TINK_SERVER_IMAGE_NAME) MANIFEST_TAG=$(TINK_CONTROLLER_IMAGE_TAG)
	$(MAKE) set-server-manifest-image MANIFEST_IMG=$(REGISTRY)/$(TINK_SERVER_IMAGE_NAME) MANIFEST_TAG=$(TINK_SERVER_IMAGE_TAG)
	$(MAKE) release-manifests

.PHONY: release-manifests ## Builds the manifests to publish with a release.
release-manifests: bin/kustomize $(RELEASE_DIR)
	kustomize build config/default > $(RELEASE_DIR)/tink.yaml

.PHONY: clean-release
clean-release: ## Remove the release folder
	rm -rf $(RELEASE_DIR)
