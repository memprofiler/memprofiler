deps:
	go get -u -v golang.org/x/tools/cmd/stringer
	go get -u -v github.com/golang/protobuf/protoc-gen-go

schema:
	protoc -I schema schema/memprofiler.proto --go_out=plugins=grpc:schema

generate:
	go generate ./...

build: schema generate
	go build -o memprofiler github.com/vitalyisaev2/memprofiler/server

.PHONY: schema
