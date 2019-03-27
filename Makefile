update:
	go get changes
	go mod tidy

tools:
	go get -u -v github.com/golangci/golangci-lint/cmd/golangci-lint

generate:
	protoc -I schema schema/*.proto  --go_out=plugins=grpc:schema
	go generate ./...

build:
	go build -o memprofiler github.com/memprofiler/memprofiler/server

run:
	./memprofiler -c ./server/config/example.yaml

lint:
	golangci-lint run --enable-all ./...

test:
	overalls -project=github.com/memprofiler/memprofiler -covermode=count -ignore=test,misc,vendor -concurrency=2
	go tool cover -func=./overalls.coverprofile

integration_test:
	go test -c ./test -o memprofiler-test && ./memprofiler-test -test.count=1

.PHONY: schema test
