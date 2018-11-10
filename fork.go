package pipe

import (
	"context"
)

//// RCopyPipe writes value Var to its stream
//type RCopyPipe struct {
//	Var interface{}
//}
//
//func (p *RCopyPipe) Go(ctx context.Context, stream Stream) error {
//	return stream.Write(p.Var)
//}
//
//// WBufferPipe reads from its stream, buffering values until the stream is closed
//// and then writes all the values as one list to the given output stream.
//type WBufferPipe struct {
//	To Stream
//}
//
//func (p *WBufferPipe) Go(ctx context.Context, stream Stream) error {
//	var values []interface{}
//	for {
//		v, err := stream.Read()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		values = append(values, v)
//	}
//	if len(values) == 0 {
//		return nil
//	}
//
//	return p.To.Write(values)
//}

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

//// ForkPipe runs a sub-pipe for each input value of the stream.
//// Each output of the pipe is collected
//// as a slice of interfaces and forwarded to the next module in the main pipe.
//type ForkPipe struct {
//	modules []Pipe
//}
//
//func (p *ForkPipe) Go(ctx context.Context, stream Stream) error {
//	var modules = make([]Pipe, len(p.modules) + 2)
//	copy(modules[1:], p.modules)
//	r, w := &RCopyPipe{}, &WBufferPipe{To: stream}
//	modules[0] = r
//	modules[len(modules)-1] = w
//
//	for {
//		v, err := stream.Read()
//		if err != nil {
//			return err
//		}
//		r.Var = v
//		err = Run(ctx, modules)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
