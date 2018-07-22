package filesystem

import (
	"encoding/json"
	"io"
)

// codec is responsible for serialization-deserialization routines
type codec interface {
	encode(io.Writer, interface{}) error
	decode(io.Reader, interface{}) error
}

// jsonCodec represents structs in JSON format
type jsonCodec struct{}

const (
	jsonPrefix = ""
	jsonIndent = "    "
)

func (*jsonCodec) encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent(jsonPrefix, jsonIndent)
	return encoder.Encode(v)
}

func (*jsonCodec) decode(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}
