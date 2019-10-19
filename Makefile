server := rover-linux-x86_64
cli := cmd/rover/rover-linux-x86_64
binaries := ${server} ${cli}
all: ${binaries}

.PHONY: server ${binaries} cli test
server: ${server}
cli: ${cli}

${server}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

${cli}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build server cli
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ${TEST_ARGS} ./...
