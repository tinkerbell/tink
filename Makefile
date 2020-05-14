server := cmd/tink-server
cli := cmd/tink-cli
worker := cmd/tink-worker
binaries := ${server} ${cli} ${worker}
all: ${binaries}

.PHONY: server ${binaries} cli worker test
server: ${server}
cli: ${cli}
worker : ${worker}

${server} ${cli} ${worker}:
	CGO_ENABLED=0 go build -o $@ ./$@

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build tinkerbell boots
test:
	go clean -testcache
	go test ./test -v
