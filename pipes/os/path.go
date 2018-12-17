package os

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"os"
	"path/filepath"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "path",
		Constructor: func(command *console.Command) pipe.Pipe {
			return PathPipe{
				where: command.Arg(0).Template(),
			}
		},
	})
}

type PathPipe struct {
	where *tap.Template
}

func (PathPipe) setup(stream pipe.Stream) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		f := tap.OpenFileInfo(path, info)
		return stream.Write(nil, f)
	}
}

func (p PathPipe) Go(ctx context.Context, stream pipe.Stream) error {
	fn := p.setup(stream)
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		path, err := (*p.where).Render(f.Context())
		if err != nil {
			return err
		}

		err = filepath.Walk(path, fn)
		if err != nil {
			return err
		}
	}
}
