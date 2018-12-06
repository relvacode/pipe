package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "print",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &PrintPipe{
				Template: console.Any().Template(),
			}
		},
	})
}

type PrintPipe struct {
	Template *tap.Template
}

func (p *PrintPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		v, err := p.Template.Render(f.Context())
		if err != nil {
			return err
		}

		err = stream.Write(nil, v)
		if err != nil {
			return err
		}
	}
}
