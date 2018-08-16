package web

import (
	"encoding/json"
	"io"
)

const (
	prefix = ""
	indent = "    "
)

// newJSONEncoder makes new encoder that creates pretty-printed JSON files
func newJSONEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	encoder.SetIndent(prefix, indent)
	return encoder
}
