package wire

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
)

func init() {
	Define(`yaml`, func() Protocol {
		return YAMLProtocol{}
	})
}

type YAMLProtocol struct {
}

func (YAMLProtocol) Encode(w io.Writer) Encoder {
	return func(x interface{}) error {
		b, err := yaml.Marshal(x)
		if err != nil {
			return err
		}
		r := bytes.NewReader(b)
		_, err = io.Copy(w, r)
		return err
	}
}

func (YAMLProtocol) Decode(r io.Reader) Decoder {
	var b bytes.Buffer
	return func() (interface{}, error) {
		defer b.Reset()

		n, err := io.Copy(&b, r)
		if n == 0 {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}

		var x interface{}
		err = yaml.Unmarshal(b.Bytes(), &x)
		return x, err
	}
}
