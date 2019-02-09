package filesystem

import (
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// codec is responsible for serialization-deserialization routines
type codec interface {
	encode(io.Writer, proto.Message) error
	decode(io.Reader, proto.Message) error
}

// jsonCodec represents structs in JSON format
type jsonCodec struct {
	marshaller *jsonpb.Marshaler
}

func (c *jsonCodec) encode(w io.Writer, v proto.Message) error {
	return c.marshaller.Marshal(w, v)
}

func (c *jsonCodec) decode(r io.Reader, v proto.Message) error {
	return jsonpb.Unmarshal(r, v)
}

func newJSONCodec() codec {
	return &jsonCodec{
		marshaller: &jsonpb.Marshaler{
			EnumsAsInts:  true,
			EmitDefaults: true,
		},
	}
}
