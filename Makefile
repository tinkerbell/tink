server := rover-server
cli := rover-cli
worker := rover-worker
binaries := ${server} ${cli} ${worker}
all: ${binaries}

.PHONY: server ${binaries} cli worker test
server: ${server}
cli: ${cli}
worker : ${worker}

${server}:
	CGO_ENABLED=0 go build -o $@ .

${cli}:
	CGO_ENABLED=0 go build -o ./cmd/rover/$@ ./cmd/rover

${worker}:
	CGO_ENABLED=0 go build -o ./worker/$@ ./worker/

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build server cli
test:
	go clean -testcache
	go test ./test -v
