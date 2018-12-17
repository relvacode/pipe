package wire

import (
	"encoding/json"
	"io"
)

func init() {
	Define(`json`, func() Protocol {
		return JSONProtocol{}
	})
}

type JSONProtocol struct {
}

func (JSONProtocol) Encode(w io.Writer) Encoder {
	var e = json.NewEncoder(w)
	e.SetIndent("", "  ")

	return func(x interface{}) error {
		return e.Encode(x)
	}
}

func (JSONProtocol) Decode(r io.Reader) Decoder {
	e := json.NewDecoder(r)
	return func() (interface{}, error) {
		var x interface{}
		err := e.Decode(&x)
		return x, err
	}
}
