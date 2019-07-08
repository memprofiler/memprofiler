package tsdb

import (
	"encoding/base64"

	"github.com/golang/protobuf/proto"
)

// codec is responsible for serialization-deserialization routines
type codec interface {
	encode(proto.Message) (string, error)
	decode(string, proto.Message) error
}

// b64Codec represents structs in base64 format
type b64Codec struct{}

func (c *b64Codec) encode(v proto.Message) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(v.String())), nil
}

func (c *b64Codec) decode(v string, p proto.Message) error {
	decodedStr, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err
	}
	return proto.UnmarshalText(string(decodedStr), p)
}

func newB64Codec() codec {
	return &b64Codec{}
}
