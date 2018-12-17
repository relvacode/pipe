package os

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/rjeczalik/notify"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "watch",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &WatchPipe{
				path: console.Arg(0).String(),
			}
		},
	})
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
