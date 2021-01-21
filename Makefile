server := cmd/tink-server
cli := cmd/tink-cli
worker := cmd/tink-worker
binaries := $(server) $(cli) $(worker)

version := $(shell git rev-parse --short HEAD)
tag := $(shell git tag --points-at HEAD)
ifneq (,$(tag))
version := $(tag)-$(version)
endif
LDFLAGS := -ldflags "-X main.version=$(version)"
export CGO_ENABLED := 0

all: $(binaries)

.PHONY: server $(binaries) cli worker test
server: $(server)
cli: $(cli)
worker : $(worker)

$(server) $(cli) $(worker):
	go build $(LDFLAGS) -o $@ ./$@

run: $(binaries)
	docker-compose up -d --build db
	docker-compose up --build tinkerbell boots
test:
	go clean -testcache
	go test ./... -v

verify:
	goimports -d .
	golint ./...
