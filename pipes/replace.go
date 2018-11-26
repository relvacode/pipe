package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/relvacode/rewrite"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "replace",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &ReplacePipe{
				What: console.Arg(0).Template(),
				With: console.Arg(0).Template(),
			}
		},
	})
}

type ReplacePipe struct {
	What *tap.Template
	With *tap.Template
}

func (p *ReplacePipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		wf, err := p.What.Render(f.Context())
		if err != nil {
			return err
		}
		if wf == "" {
			return errors.New("cannot search for empty string in replace")
		}

		wd, err := p.With.Render(f.Context())
		if err != nil {
			return err
		}

		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		replacer := rewrite.New(r, []byte(wf), []byte(wd))
		err = stream.Write(tap.ReadProxyCloser(replacer, r))
		if err != nil {
			return err
		}
	}
}
