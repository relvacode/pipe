package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/valve"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "print",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return &PrintPipe{
				Template: valve.All().String(),
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
