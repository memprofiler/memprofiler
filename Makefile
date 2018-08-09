deps:
	go get -u -v -d google.golang.org/grpc
	go get -u -v github.com/golang/protobuf/...
	go get -u -v golang.org/x/tools/cmd/stringer

generate:
	protoc -I schema schema/memprofiler.proto --go_out=plugins=grpc:schema
	go generate ./...

build:
	go build -o memprofiler github.com/vitalyisaev2/memprofiler/server

.PHONY: schema
