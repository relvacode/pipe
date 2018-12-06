package pipes

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"reflect"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "skip",
		Constructor: func(console *console.Command) pipe.Pipe {
			return SkipPipe{
				Skip: console.Arg(0).Int(),
			}
		},
	})
	pipe.Define(pipe.Pkg{
		Name: "limit",
		Constructor: func(console *console.Command) pipe.Pipe {
			return LimitPipe{
				Limit: console.Arg(0).Int(),
			}
		},
	})
	pipe.Define(pipe.Pkg{
		Name: "flatten",
		Constructor: func(console *console.Command) pipe.Pipe {
			return FlattenPipe{}
		},
	})
}

type SkipPipe struct {
	Skip *int64
}

func (p SkipPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := int64(0); ; i++ {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}
		if i < *p.Skip {
			continue
		}

		err = stream.With(f).Write(nil, f.Object)
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
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		if i < *p.Limit {
			err = stream.Write(nil, f.Object)
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
		err := stream.Write(nil, v.Index(i).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p FlattenPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		v := reflect.ValueOf(f.Object)
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			err = p.each(v, stream)
		default:
			err = stream.Write(nil, f.Object)
		}

		if err != nil {
			return err
		}
	}
}
