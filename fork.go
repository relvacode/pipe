package pipe

import (
	"context"
	"io"
)

type WFramePipe struct {
	Frame *DataFrame
}

func (p *WFramePipe) Go(ctx context.Context, stream Stream) error {
	return stream.With(p.Frame).Write(p.Frame.Object)
}

// WBufferPipe reads from its stream, buffering values until the stream is closed
// and then writes all the values as one list to the given output stream.
type WBufferPipe struct {
	To Stream
}

func (p *WBufferPipe) Go(ctx context.Context, stream Stream) error {
	var values []interface{}
	for {
		f, err := stream.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		values = append(values, f.Object)
	}
	if len(values) == 0 {
		return nil
	}

	return p.To.Write(values)
}

type RCopyPipe struct {
	From Stream
}

func (p *RCopyPipe) Go(ctx context.Context, stream Stream) error {
	for {
		f, err := p.From.Read()
		if err != nil {
			return err
		}

		err = stream.With(f).Write(f.Object)
		if err != nil {
			return err
		}
	}
}

type WCopyPipe struct {
	To Stream
}

func (p *WCopyPipe) Go(ctx context.Context, stream Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		err = p.To.With(f).Write(f.Object)
		if err != nil {
			return err
		}
	}
}

type SubPipe []Runnable

func (p SubPipe) setup(stream Stream) []Runnable {
	pipes := make([]Runnable, len(p)+2)
	copy(pipes[1:len(pipes)-1], p)

	pipes[0] = Runnable{
		Pipe: &RCopyPipe{
			From: stream,
		},
	}
	pipes[len(pipes)-1] = Runnable{
		Pipe: &WCopyPipe{
			To: stream,
		},
	}

	return pipes
}

func (p SubPipe) Go(ctx context.Context, stream Stream) error {
	return Run(ctx, p.setup(stream))
}

type ForkPipe []Runnable

func (p ForkPipe) Go(ctx context.Context, stream Stream) error {
	var modules = make([]Runnable, len(p)+2)
	copy(modules[1:], p)
	r, w := &WFramePipe{}, &WBufferPipe{To: stream}
	modules[0] = Runnable{Pipe: r}
	modules[len(modules)-1] = Runnable{Pipe: w}

	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}
		r.Frame = f
		err = Run(ctx, modules).ErrorOrNil()
		if err != nil {
			return err
		}
	}
}
