package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "open",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &OpenPipe{
				files: console.Shell(),
			}
		},
	})
}

type File struct {
	Name string
	Path string
	Size int64
	File *os.File
}

func (f *File) Read(b []byte) (int, error) {
	return f.File.Read(b)
}

func (f *File) Close() error {
	return f.File.Close()
}

// OpenPipe opens files matching one or more glob expressions.
// It sends each file handle to the next module.
type OpenPipe struct {
	files *[]string
}

func (p *OpenPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for _, n := range *p.files {
		files, err := filepath.Glob(n)
		if err != nil {
			return errors.Wrapf(err, "glob %q", n)
		}
		logrus.Debugf("%d files found matching pattern %s", len(files), n)
		for _, fn := range files {
			f, err := os.Open(fn)
			if err != nil {
				return err
			}
			i, err := f.Stat()
			if err != nil {
				return err
			}

			if i.IsDir() {
				return errors.Errorf("cannot open %q: is directory", fn)
			}

			err = stream.Write(&File{
				Name: f.Name(),
				Path: fn,
				Size: i.Size(),
				File: f,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}
