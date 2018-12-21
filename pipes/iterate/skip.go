package iterate

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "skip",
		Constructor: func(console *console.Command) pipe.Pipe {
			return SkipPipe{
				Skip: console.Arg(0).Int(),
			}
		},
	})
}

type SkipPipe struct {
	Skip *int64
}

func (p SkipPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := int64(0); ; i++ {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}
		if i < *p.Skip {
			continue
		}

		err = stream.With(f).Write(nil, f.Object)
		if err != nil {
			return err
		}
	}
}
