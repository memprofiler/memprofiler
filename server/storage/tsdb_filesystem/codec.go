package filesystem

import (
	"bytes"
	"encoding/base64"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// codec is responsible for serialization-deserialization routines
type codec interface {
	encode(proto.Message) (string, error)
	decode(string, proto.Message) error
}

// jsonCodec represents structs in JSON format
type jsonCodec struct {
	marshaller *jsonpb.Marshaler
}

func (c *jsonCodec) encode(v proto.Message) (string, error) {
	b := bytes.Buffer{}
	err := c.marshaller.Marshal(&b, v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func (c *jsonCodec) decode(v string, p proto.Message) error {
	by, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err
	}
	b := bytes.Buffer{}
	b.Write(by)
	return jsonpb.Unmarshal(&b, p)
}

func newJSONCodec() codec {
	return &jsonCodec{
		marshaller: &jsonpb.Marshaler{
			EnumsAsInts:  true,
			EmitDefaults: true,
		},
	}
}
