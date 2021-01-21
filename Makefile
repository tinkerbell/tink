# Only use the recipes defined in these makefiles
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:
# Delete target files if there's an error
# This avoids a failure to then skip building on next run if the output is created by shell redirection for example
# Not really necessary for now, but just good to have already if it becomes necessary later.
.DELETE_ON_ERROR:
# Treat the whole recipe as a one shell script/invocation instead of one-per-line 
.ONESHELL:
# Use bash instead of plain sh 
SHELL := bash
.SHELLFLAGS := -o pipefail -euc

server := cmd/tink-server/tink-server
cli := cmd/tink-cli/tink-cli
worker := cmd/tink-worker/tink-worker
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
	go build $(LDFLAGS) -o $@ ./$(@D)

crossbinaries := $(addsuffix -linux-,$(binaries))
crossbinaries := $(crossbinaries:=386) $(crossbinaries:=amd64) $(crossbinaries:=arm64) $(crossbinaries:=armv6) $(crossbinaries:=armv7)
crosscompile: $(crossbinaries)
.PHONY: crosscompile $(crossbinaries)

%-386:   FLAGS=GOARCH=386
%-amd64: FLAGS=GOARCH=amd64
%-arm64: FLAGS=GOARCH=arm64
%-armv6: FLAGS=GOARCH=arm GOARM=6
%-armv7: FLAGS=GOARCH=arm GOARM=7
$(crossbinaries):
	$(FLAGS) GOOS=linux go build $(LDFLAGS) -o $@ ./$(@D)

run: $(binaries)
	docker-compose up -d --build db
	docker-compose up --build tinkerbell boots
test:
	go clean -testcache
	go test ./... -v

verify:
	goimports -d .
	golint ./...
