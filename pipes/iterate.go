package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/valve"
	"reflect"
	"time"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "skip",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return SkipPipe{
				Skip: valve.All().Int(),
			}
		},
	})
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "limit",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return LimitPipe{
				Limit: valve.All().Int(),
			}
		},
	})
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "flatten",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return FlattenPipe{}
		},
	})

	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "delay",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return DelayPipe{
				Template: valve.All().String(),
			}
		},
	})

	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "every",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return EveryPipe{
				Duration: valve.All().Duration(),
			}
		},
	})
}

type SkipPipe struct {
	Skip *int64
}

func (p SkipPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := int64(0); ; i++ {
		f, err := stream.Read()
		if err != nil {
			return err
		}
		if i < *p.Skip {
			continue
		}

		err = stream.With(f).Write(f.Object)
		if err != nil {
			return err
		}
	}
}

type LimitPipe struct {
	Limit *int64
}

func (p LimitPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := int64(0); ; i++ {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		if i < *p.Limit {
			err = stream.Write(f.Object)
			if err != nil {
				return err
			}
		}
	}
}

// FlattenPipe takes an array type and flattens it out
// by sending each item in the array as a separate item to the output stream
type FlattenPipe struct {
}

func (FlattenPipe) each(v reflect.Value, stream pipe.Stream) error {
	for i := 0; i < v.Len(); i++ {
		err := stream.Write(v.Index(i).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p FlattenPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		v := reflect.ValueOf(f.Object)
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			err = p.each(v, stream)
		default:
			err = stream.Write(f.Object)
		}

		if err != nil {
			return err
		}
	}
}

type DelayPipe struct {
	Template *string
}

func (p DelayPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		s, err := f.Var(*p.Template)
		if err != nil {
			return err
		}

		d, err := time.ParseDuration(s)
		if err != nil {
			return err
		}

		t := time.After(d)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t:
			err = stream.Write(f.Object)
			if err != nil {
				return err
			}
		}
	}
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
			err := stream.Write(t)
			if err != nil {
				return err
			}
		}
	}
}
