server := rover-linux-x86_64
cli := cmd/rover/rover-linux-x86_64
worker := worker/rover-worker-linux-x86_64
binaries := ${server} ${cli} ${worker}
GOPATH := $(shell go env GOPATH)
all: proto-gen ${binaries}

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

proto-gen:
	protoc protos/template/template.proto --go_out=plugins=grpc:$(GOPATH)/src
	protoc protos/target/target.proto --go_out=plugins=grpc:$(GOPATH)/src
	protoc protos/workflow/workflow.proto --go_out=plugins=grpc:$(GOPATH)/src
	protoc -I$(GOPATH)/src --go_out=plugins=grpc:$(GOPATH)/src $(GOPATH)/src/github.com/packethost/rover/protos/rover/rover.proto

run: ${binaries}
	docker-compose up -d --build db
	docker-compose up --build server cli
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ${TEST_ARGS} ./...
