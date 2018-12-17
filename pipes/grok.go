package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/vjeantet/grok"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "grok",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &GrokPipe{
				pattern: console.Any().Template(),
			}
		},
	})
}

type GrokPipe struct {
	pattern *tap.Template
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

		pattern, err := (*p.pattern).Render(f.Context())
		if err != nil {
			return err
		}

		values, err := g.ParseTyped(pattern, str)
		if err != nil {
			return err
		}

		err = stream.Write(nil, values)
		if err != nil {
			return err
		}
	}
}
