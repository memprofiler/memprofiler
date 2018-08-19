package web

import (
	"github.com/gogo/protobuf/jsonpb"
)

const (
	indent = "    "
)

// newJSONMarshaler makes new marshaler that creates pretty-printed JSON files
func newJSONMarshaler() *jsonpb.Marshaler {
	marshaler := &jsonpb.Marshaler{
		Indent:       indent,
		EmitDefaults: true,
	}
	return marshaler
}
