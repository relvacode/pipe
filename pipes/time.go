package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"time"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "every",
		Constructor: func(console *console.Command) pipe.Pipe {
			return EveryPipe{
				Duration: console.Arg(0).Duration(),
			}
		},
	})
	pipe.Define(pipe.Pkg{
		Name: "timeout",
		Constructor: func(console *console.Command) pipe.Pipe {
			return TimeoutPipe{
				Duration: console.Arg(0).Duration(),
			}
		},
	})
	pipe.Define(pipe.Pkg{
		Name: "delay",
		Constructor: func(console *console.Command) pipe.Pipe {
			return DelayPipe{
				Duration: console.Arg(0).Duration(),
			}
		},
	})
}

type EveryPipe struct {
	Duration *time.Duration
}

func (p EveryPipe) Go(ctx context.Context, stream pipe.Stream) error {
	ticker := time.NewTicker(*p.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t := <-ticker.C:
			err := stream.Write(nil, t)
			if err != nil {
				return err
			}
		}
	}
}

type TimeoutPipe struct {
	Duration *time.Duration
}

func (p TimeoutPipe) Go(ctx context.Context, stream pipe.Stream) error {
	timeout, cancel := context.WithTimeout(context.Background(), *p.Duration)

	for {
		f, err := stream.Read(timeout.Done())
		cancel()
		if err != nil {
			return err
		}
		err = stream.Write(nil, f.Object)
		if err != nil {
			return err
		}

		timeout, cancel = context.WithTimeout(context.Background(), *p.Duration)
	}
}

type DelayPipe struct {
	Duration *time.Duration
}

func (p DelayPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		t := time.After(*p.Duration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t:
			err = stream.Write(nil, f.Object)
			if err != nil {
				return err
			}
		}
	}
}
