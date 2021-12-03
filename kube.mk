PHONY: generate
generate: generate-go generate-manifests # Generate code, manifests etc.

.PHONY: generate-go
generate-go: bin/controller-gen bin/gofumpt # Generate Go code.
	controller-gen object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./pkg/apis/..."
	gofumpt -w -s ./pkg/apis

.PHONY: generate-manifests
generate-manifests: bin/controller-gen # Generate manifests e.g. CRD, RBAC etc.
	controller-gen \
		paths=./pkg/apis/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook
	prettier --write ./config/crd/
