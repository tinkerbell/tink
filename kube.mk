## --------------------------------------
## Generate
## --------------------------------------

.PHONY: generate
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

TINK_SERVER_IMAGE ?= quay.io/tinkerbell/tink-server
TINK_CONTROLLER_IMAGE ?= quay.io/tinkerbell/tink-controller
TINK_SERVER_TAG ?= latest
TINK_CONTROLLER_TAG ?= latest
NAMESPACE ?= tink-system

out/release/default/kustomization.yaml: config/default/kustomization.yaml
	rm -rf out/
	mkdir -p out/
	cp -a config/ out/release/

out/release/tink.yaml: bin/kustomize generate-manifests out/release/default/kustomization.yaml
	(cd out/release/default && \
		kustomize edit set image server=$(TINK_SERVER_IMAGE):$(TINK_CONTROLLER_TAG) controller=$(TINK_CONTROLLER_IMAGE):$(TINK_CONTROLLER_TAG) && \
		kustomize edit set namespace $(NAMESPACE) \
	)
	kustomize build out/release/default -o $@
	prettier --write $@

release-manifests: out/release/tink.yaml ## Builds the manifests to publish with a release.
