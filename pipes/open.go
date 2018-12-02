package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"path/filepath"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "open",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &OpenPipe{
				glob: console.Arg(0).Template(),
			}
		},
	})
}

// OpenPipe opens files matching one or more glob expressions.
// It sends each file handle to the next module.
type OpenPipe struct {
	glob *tap.Template
}

func (p *OpenPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		pattern, err := p.glob.Render(f.Context())
		if err != nil {
			return err
		}

		files, err := filepath.Glob(pattern)
		if err != nil {
			return errors.Wrapf(err, "glob %q", pattern)
		}
		for _, fn := range files {
			f, err := tap.OpenFile(fn)
			if err != nil {
				return errors.Wrapf(err, "open %q", fn)
			}

			err = stream.Write(f)
			if err != nil {
				return err
			}
		}
	}

}
