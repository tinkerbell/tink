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
