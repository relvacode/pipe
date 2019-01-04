package aggregate

import (
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"io"
)

// Aggregation is a pipe that collects all values and emits one value.
type Aggregation interface {
	// Each is called for each frame received on the input stream.
	Each(interface{}) error
	// Final is called once all frames have been received.
	// It returns the final object aggregating all Of the above values.
	Final() (interface{}, error)
}

func NewAggregator(command *console.Command, f func() Aggregation) *Pipe {
	return &Pipe{
		Of:   command.Any().Expression(),
		Init: f,
	}
}

type Pipe struct {
	Of   *console.Expression
	Init func() Aggregation
}

func (p Pipe) Go(ctx context.Context, stream pipe.Stream) error {
	var ag = p.Init()
each:
	for {
		f, err := stream.Read(nil)
		switch err {
		case io.EOF:
			break each
		case nil:
			v, err := (*p.Of).Eval(f.Context())
			if err != nil {
				return err
			}

			err = ag.Each(v)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	v, err := ag.Final()
	if err != nil {
		return err
	}

	return stream.With(nil).Write(nil, v)
}
