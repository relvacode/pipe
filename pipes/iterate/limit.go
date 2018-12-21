package iterate

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "limit",
		Constructor: func(console *console.Command) pipe.Pipe {
			return LimitPipe{
				Limit: console.Arg(0).Int(),
			}
		},
	})
}

type LimitPipe struct {
	Limit *int64
}

func (p LimitPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := int64(0); ; i++ {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		if i < *p.Limit {
			err = stream.Write(nil, f.Object)
			if err != nil {
				return err
			}
		}
	}
}
