package encoding

import (
	"bytes"
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
)

type Decoder func() (interface{}, error)

type Encoder func(interface{}) error

// A Protocol is an interface that implements automatic encoding and decoding of values
type Protocol interface {
	Decode(r io.Reader) Decoder
	Encode(w io.Writer) Encoder
}

func Define(name string, p func() Protocol) {
	pipe.Define(pipe.Pkg{
		Name: name,
		Constructor: func(_ *console.Command) pipe.Pipe {
			return &Pipe{
				Protocol: p(),
			}
		},
	})
}

type Pipe struct {
	Protocol Protocol
}

func (p Pipe) DecodeProtocol(r io.Reader, stream pipe.Stream) error {
	decoder := p.Protocol.Decode(r)
	for {
		x, err := decoder()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = stream.Write(nil, x)
		if err != nil {
			return err
		}
	}
}

func (p Pipe) EncodeProtocol(x interface{}, stream pipe.Stream) error {
	var b bytes.Buffer
	var e = p.Protocol.Encode(&b)
	err := e(x)
	if err != nil {
		return err
	}

	return stream.Write(nil, &b)
}

func (p Pipe) Go(_ context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		switch x := f.Object.(type) {
		case io.Reader:
			err = p.DecodeProtocol(x, stream)
			_ = tap.Close(x)

		default:
			err = p.EncodeProtocol(x, stream)
		}

		if err != nil {
			return err
		}
	}
}
