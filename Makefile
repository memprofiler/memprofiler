deps:
	go get -u -v github.com/deckarep/golang-set
	go get -u -v github.com/sirupsen/logrus
	go get -u -v google.golang.org/grpc
	go get -u -v github.com/golang/protobuf/...
	go get -u -v github.com/improbable-eng/grpc-web/go/grpcweb
	go get -u -v golang.org/x/tools/cmd/stringer
	go get -u -v gopkg.in/yaml.v2
	go get -u -v gonum.org/v1/gonum/stat
	go get -u -v github.com/stretchr/testify/mock

generate:
	protoc -I schema schema/*.proto  --go_out=plugins=grpc:schema
	go generate ./...

build:
	go build -o memprofiler github.com/memprofiler/memprofiler/server

run:
	./memprofiler -c ./server/config/example.yaml

test:
	go test -v ./...

.PHONY: schema
