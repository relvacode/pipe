package pipe

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

type Runnable struct {
	Pipe Pipe
	Tag  *Tag
}

func (r Runnable) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%T", r.Pipe)
	if r.Tag != nil {
		fmt.Fprintf(&b, " as %s", *r.Tag)
	}
	return b.String()
}

func RunIO(ctx context.Context, input Pipe, modules []Runnable, output Pipe) error {
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

func Run(ctx context.Context, modules []Runnable) error {
	logrus.Debugf("about to run %d modules", len(modules))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var streams = make([]*stream, len(modules))
	streams[len(streams)-1] = NewStream(ctx, nil)

	for i := len(modules) - 2; i > -1; i-- {
		var s = NewStream(ctx, modules[i].Tag)
		s.Down(streams[i+1])
		streams[i] = s
	}

	// Link all the streams in reverse
	for i := 1; i < len(streams); i++ {
		streams[i].Up(streams[i-1])
	}
	logrus.Debugf("streams: %v", streams)

	var errs = make(chan error, len(modules))

	for i := 0; i < len(modules); i++ {
		logrus.Debugf("starting module %s %#v", modules[i], modules[i].Pipe)
		go func(s *stream, e Pipe) {
			defer func() {
				logrus.Debugf("module %T stopped", e)
			}()
			defer s.Close()

			// Panic recovery
			defer func() {
				r := recover()
				if r == nil {
					return
				}
				errs <- errors.Errorf("%s", r)
			}()

			err := e.Go(ctx, s)
			if err != nil {
				err = errors.Wrapf(err, "stream %s in pipe %T", s, e)
			}
			errs <- err

		}(streams[i], modules[i].Pipe)
	}

	var complete *multierror.Error
	for i := 0; i < len(streams); i++ {
		err := <-errs
		cause := errors.Cause(err)
		if err == nil || cause == io.EOF || cause == context.Canceled {
			continue
		}
		cancel()
		complete = multierror.Append(complete, err)
	}
	return complete.ErrorOrNil()
}
