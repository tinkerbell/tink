server := rover-linux-x86_64
cli := cmd/rover/rover-linux-x86_64
worker := worker/rover-worker-linux-x86_64
binaries := ${server} ${cli} ${worker}
all: ${binaries}

.PHONY: server ${binaries} cli worker test
server: ${server}
cli: ${cli}
worker : ${worker}

${server}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

${cli}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

${worker}:
	CGO_ENABLED=0 GOOS=linux go build -o $@ ./$(@D)

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build server cli
test:
	go clean -testcache
	go test ./test -v
