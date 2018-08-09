deps:
	go get -u -v -d google.golang.org/grpc
	go get -u -v -d github.com/golang/protobuf/...

	go get -u -v golang.org/x/tools/cmd/stringer

schema:
	protoc -I schema schema/memprofiler.proto --go_out=plugins=grpc:schema

generate:
	go generate ./...

build: schema generate
	go build -o memprofiler github.com/vitalyisaev2/memprofiler/server

.PHONY: schema
