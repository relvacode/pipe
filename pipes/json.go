package pipes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "json",
		Family: []pipe.Pkg{
			{
				Name: "decode",
				Constructor: func(command *console.Command) pipe.Pipe {
					return JSONDecodePipe{}
				},
			},
			{
				Name: "encode",
				Constructor: func(command *console.Command) pipe.Pipe {
					return &JSONEncodePipe{
						what: command.Any().Default(console.DefaultExpression{}).Expression(),
					}
				},
			},
		},
	})
}

type JSONDecodePipe struct {
}

func (p JSONDecodePipe) Go(_ context.Context, stream pipe.Stream) error {
	for {
		v, err := stream.Read(nil)
		if err != nil {
			return err
		}

		r, err := tap.Reader(v.Object)
		if err != nil {
			return err
		}

		decoder := json.NewDecoder(r)
		var i int
		for ; ; i++ {
			var x interface{}
			err := decoder.Decode(&x)
			if err != nil {
				tap.Close(r)
				if err == io.EOF {
					break
				}
				return err
			}
			err = stream.Write(nil, x)
			if err != nil {
				tap.Close(r)
				return err
			}
		}
		if i == 0 {
			return errors.New("json: no data in stream")
		}
	}
}

type JSONEncodePipe struct {
	what *console.Expression
}

func (p JSONEncodePipe) Go(_ context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		v, err := (*p.what).Eval(f.Context())
		if err != nil {
			return err
		}

		var buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(v)
		if err != nil {
			return err
		}

		err = stream.Write(nil, buf)
		if err != nil {
			return err
		}
	}
}
