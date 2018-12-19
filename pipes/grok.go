package pipes

import (
	"bufio"
	"context"
	"fmt"
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

		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		s := bufio.NewScanner(r)
		for s.Scan() {
			values, err := g.ParseTyped(*p.pattern, s.Text())
			if err != nil {
				return err
			}
			if len(values) == 0 {
				continue
			}

			err = stream.Write(nil, values)
			if err != nil {
				fmt.Println(err)
			}
		}

		err = s.Err()
		if err != nil {
			return err
		}
	}
}
