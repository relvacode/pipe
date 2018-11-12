package pipe

import (
	"context"
	"github.com/relvacode/pipe/console"
)

type Pipe interface {
	Go(context.Context, Stream) error
}

// Func is a pipe implemented as a stateless function.
// Useful for creating simple pipes.
type Func func(context.Context, Stream) error

type pipeFuncWrapper struct {
	f Func
}

func (p *pipeFuncWrapper) Go(ctx context.Context, stream Stream) error {
	return p.f(ctx, stream)
}

// Create an instance of a Pipe from a pipe function
func FromFunc(f Func) Builder {
	return func(*console.Command) Pipe {
		return &pipeFuncWrapper{
			f: f,
		}
	}
}
