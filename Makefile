schema:
	protoc -I schema schema/memprofiler.proto --go_out=plugins=grpc:schema

build: schema
