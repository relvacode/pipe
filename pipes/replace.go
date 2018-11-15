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
			args := console.SplitNArgs(2)
			return &ReplacePipe{
				What: args[0],
				With: args[1],
			}
		},
	})
}

type ReplacePipe struct {
	What *string
	With *string
}

func (p *ReplacePipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		wf, err := f.Var(*p.What)
		if err != nil {
			return err
		}
		if wf == "" {
			return errors.New("cannot search for empty string in replace")
		}

		wd, err := f.Var(*p.With)
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
