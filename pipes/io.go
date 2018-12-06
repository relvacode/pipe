package pipes

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "buffer",
		Constructor: func(command *console.Command) pipe.Pipe {
			return BufferPipe{}
		},
	})
}

type BufferPipe struct {
}

func (BufferPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}
		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, r)
		if err != nil {
			return errors.Wrap(err, "copy buffer")
		}

		err = stream.Write(nil, &b)
		if err != nil {
			return err
		}
	}
}
