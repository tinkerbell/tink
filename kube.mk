## --------------------------------------
## Generate
## --------------------------------------

CONTROLLER_GEN 	:= go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10
GOFUMPT 		:= go run mvdan.cc/gofumpt@v0.1
KUSTOMIZE 		:= go run sigs.k8s.io/kustomize/kustomize/v4@latestv4.5

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
