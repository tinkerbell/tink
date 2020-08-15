server := cmd/tink-server
cli := cmd/tink-cli
worker := cmd/tink-worker
binaries := ${server} ${cli} ${worker}

git_version?=$(shell git log -1 --format="%h")
version?=$(git_version)
release_tag?=$(shell git tag --points-at HEAD)
ifneq (,$(release_tag))
version:=$(release_tag)-$(version)
endif
LDFLAGS?=-ldflags "-X main.version=$(version)"

all: ${binaries}

.PHONY: server ${binaries} cli worker test
server: ${server}
cli: ${cli}
worker : ${worker}

${server} ${cli} ${worker}:
	CGO_ENABLED=0 GOOS=$$GOOS go build ${LDFLAGS} -o $@ ./$@

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build tinkerbell boots
test:
	go clean -testcache
	go test ./test -v

verify:
	goimports -d .
	golint ./...
