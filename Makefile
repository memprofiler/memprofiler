deps:
	GO111MODULE=on go get -u ./...
	go mod tidy

PROTOBUF_VERSION=3.7.1
PROTOBUF_DIR=/tmp/protoc-${PROTOBUF_VERSION}
GOLANGCI_LINT_VERSION=1.17.1
env:
	# download protobuf release
	wget -P /tmp https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
	if [ -d ${PROTOBUF_DIR} ]; then \
		rm -rf ${PROTOBUF_DIR}; \
	fi;
	mkdir -p ${PROTOBUF_DIR}
	unzip /tmp/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip -d ${PROTOBUF_DIR}

	# install protoc
	if [ -f /usr/local/bin/protoc ]; then \
		rm -rf /usr/local/bin/protoc; \
	fi;
	# make sure that permissions are sufficient
	mv ${PROTOBUF_DIR}/bin/* /usr/local/bin

	# install headers
	if [ -d /usr/local/include/google ]; then \
		rm -rf /usr/local/include/google; \
	fi;
	mv ${PROTOBUF_DIR}/include/* /usr/local/include

	# install plugin
	go get -u github.com/golang/protobuf/protoc-gen-go

	# install tools
	go get -u -v golang.org/x/tools/cmd/stringer
	go get -u -v github.com/favadi/protoc-go-inject-tag
	protoc-go-inject-tag -input=./schema/common.pb.go
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b /usr/local/bin v${GOLANGCI_LINT_VERSION}

generate:
	protoc -I schema schema/*.proto  --go_out=plugins=grpc:schema
	protoc-go-inject-tag -input=./schema/common.pb.go
	go generate ./...

build:
	GO111MODULE=on go build github.com/memprofiler/memprofiler

lint:
	golangci-lint run --enable-all ./...

PACKAGES = $(shell find ./ -type d -not -path '*/\.*' -not -path '*test*')
test:
	echo "mode: atomic" > coverage.txt
	$(foreach pkg,$(PACKAGES),\
		go test -race -coverprofile=coverage.out -covermode=atomic $(pkg);\
		tail -n +2 coverage.out >> coverage.txt;\
		rm coverage.out;)
	go tool cover -func=coverage.txt

integration_test:
	go test -c ./test -o memprofiler-test && ./memprofiler-test -test.count=1

run:
	./memprofiler -c ./server/config/example.yaml

.PHONY: schema test
