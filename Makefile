server := tinkerbell-server
cli := tinkerbell-cli
worker := tinkerbell-worker
binaries := ${server} ${cli} ${worker}
all: ${binaries}

.PHONY: server ${binaries} cli worker test
server: ${server}
cli: ${cli}
worker : ${worker}

${bindir}:
	mkdir -p $@/

${server}:
	CGO_ENABLED=0 go build -o $@ .

${cli}:
	CGO_ENABLED=0 go build -o ./cmd/tinkerbell/$@ ./cmd/tinkerbell

${worker}:
	CGO_ENABLED=0 go build -o ./worker/$@ ./worker/

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build tinkerbell boots
test:
	go clean -testcache
	go test ./test -v
