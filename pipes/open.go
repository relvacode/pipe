package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/rjeczalik/notify"
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
	pipe.Define(pipe.Pkg{
		Name: "watch",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &WatchPipe{
				path: console.Arg(0).String(),
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
		f, err := stream.Read(nil)
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

			err = stream.Write(nil, f)
			if err != nil {
				return err
			}
		}
	}
}

type ChangedFile struct {
	Event string
	Path  string
}

func (f ChangedFile) String() string {
	return f.Path
}

type WatchPipe struct {
	path *string
}

func (WatchPipe) eventType(e notify.Event) string {
	switch e {
	case notify.Write:
		return "write"
	case notify.Create:
		return "create"
	case notify.Remove:
		return "remove"
	case notify.Rename:
		return "rename"
	default:
		return e.String()
	}
}

func (p WatchPipe) Go(ctx context.Context, stream pipe.Stream) error {
	var (
		changes = make(chan notify.EventInfo, 100)
		err     = notify.Watch(*p.path, changes, notify.All)
	)

	if err != nil {
		return err
	}
	defer notify.Stop(changes)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-changes:
			err = stream.Write(nil, ChangedFile{
				Path:  e.Path(),
				Event: p.eventType(e.Event()),
			})
			if err != nil {
				return err
			}
		}
	}
}
