package pipe

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

type RuntimeError []error

func (e RuntimeError) Error() string {
	var s strings.Builder
	for _, err := range e {
		fmt.Fprintln(&s, "  - ", err)
	}
	return s.String()
}

func (e RuntimeError) ErrorOrNil() error {
	if e == nil || len(e) == 0 {
		return nil
	}
	return e
}

type Runnable struct {
	Pipe Pipe
	Tag  *Tag
}

func RunIO(ctx context.Context, input Pipe, modules []Runnable, output Pipe) RuntimeError {
	pipes := make([]Runnable, len(modules)+2)
	copy(pipes[1:], modules)
	pipes[0] = Runnable{
		Pipe: input,
	}
	pipes[len(pipes)-1] = Runnable{
		Pipe: output,
	}
	return Run(ctx, pipes)
}

func Run(ctx context.Context, runnables []Runnable) RuntimeError {
	logrus.Debugf("about to run %d pipes", len(runnables))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var streams = make([]*stream, len(runnables))
	streams[len(streams)-1] = NewStream(ctx, nil)

	for i := len(runnables) - 2; i > -1; i-- {
		var s = NewStream(ctx, runnables[i].Tag)
		s.Down(streams[i+1])
		streams[i] = s
	}

	// Link all the streams in reverse
	for i := 1; i < len(streams); i++ {
		streams[i].Up(streams[i-1])
	}
	logrus.Debugf("streams: %v", streams)

	var errs = make(chan error, len(runnables))

	for i := 0; i < len(runnables); i++ {
		go func(s *stream, e Runnable) {
			defer func() {
				logrus.Debugf("pipe %T stopped", e.Pipe)
			}()
			defer s.Close()

			logrus.Debugf("pipe %T started on stream %s", e.Pipe, s)
			err := e.Pipe.Go(s.ctx, s)
			if err != nil {
				err = errors.Wrapf(err, "%T on %s", e.Pipe, s)
			}
			errs <- err

		}(streams[i], runnables[i])
	}

	var complete RuntimeError
	for i := 0; i < len(streams); i++ {
		err := <-errs
		cause := errors.Cause(err)
		if err == nil || cause == io.EOF || cause == context.Canceled {
			continue
		}
		cancel()
		complete = append(complete, err)
	}
	if len(complete) == 0 {
		return nil
	}
	return complete
}
