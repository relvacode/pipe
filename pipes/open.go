package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "open",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &OpenPipe{
				files: console.Split(),
			}
		},
	})
}

// OpenPipe opens files matching one or more glob expressions.
// It sends each file handle to the next module.
type OpenPipe struct {
	files *[]string
}

func (p *OpenPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		_, err := stream.Read()
		if err != nil {
			return err
		}

		for _, n := range *p.files {
			files, err := filepath.Glob(n)
			if err != nil {
				return errors.Wrapf(err, "glob %q", n)
			}
			logrus.Debugf("%d files found matching pattern %s", len(files), n)
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

}
