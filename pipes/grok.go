package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/vjeantet/grok"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "grok",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &GrokPipe{
				pattern: console.Any().String(),
			}
		},
	})
}

type GrokPipe struct {
	pattern *string
}

func (p GrokPipe) Go(ctx context.Context, stream pipe.Stream) error {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		return err
	}
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		str, err := f.AsString()
		if err != nil {
			return err
		}

		values, err := g.Parse(*p.pattern, str)
		if err != nil {
			return err
		}

		err = stream.Write(nil, values)
		if err != nil {
			return err
		}
	}
}
