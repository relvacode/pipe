package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "print",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &PrintPipe{
				Template: console.String(),
			}
		},
	})
}

type PrintPipe struct {
	Template *string
}

func (p *PrintPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		v, err := f.Var(*p.Template)
		if err != nil {
			return err
		}

		err = stream.Write(v)
		if err != nil {
			return err
		}
	}
}
