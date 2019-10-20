cli := cmd/rover/rover-linux-x86_64
binaries := ${cli}
all: ${binaries}

.PHONY: ${binaries} cli test
cli: ${cli}

${cli}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

run: ${binaries}
	docker-compose up --build cli
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ${TEST_ARGS} ./...
build:
	docker build -t rover-cli ./cmd/rover
