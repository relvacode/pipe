package pipe

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"sync/atomic"
)

type Stream interface {
	// Read a frame from the stream
	Read() (*DataFrame, error)
	// Write an object back to the stream.
	// If a write happens after a read
	// then all writes after that read contain the originally read object in the frame stack
	Write(interface{}) error

	// Copy this stream so that all subsequent writes use this data frame in the stack.
	With(*DataFrame) Stream
}

var streamIds *uint64

func init() {
	var ids uint64
	streamIds = &ids
}

func NewStream(ctx context.Context, tag *Tag) *stream {
	return &stream{
		id:    atomic.AddUint64(streamIds, 1),
		ctx:   ctx,
		tag:   tag,
		ok:    make(chan struct{}),
		input: make(chan *DataFrame),
	}
}

type stream struct {
	id   uint64
	ctx  context.Context
	tag  *Tag
	f    *DataFrame
	r, w int

	input chan *DataFrame
	ok    chan struct{} // closed when downstream is closed

	up   *stream
	down *stream
}

func (s *stream) Close() {
	logrus.Debugf("%s terminated read:%d write:%d", s, s.r, s.w)
	if s.up != nil {
		close(s.up.ok)
	}
	if s.down != nil {
		close(s.down.input)
	}
}

func (s *stream) String() string {
	return fmt.Sprintf("Stream(%d: %s)", s.id, s.tag)
}

func (s *stream) Up(p *stream) {
	s.up = p
}

func (s *stream) Down(n *stream) {
	s.down = n
}

func (s *stream) Read() (*DataFrame, error) {
	select {
	case x, ok := <-s.input:
		if !ok {
			return nil, io.EOF
		}
		s.f = x
		s.r++
		logrus.Debugf("%s read %s (%d)", s, s.f, s.r)
		return x, nil
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	}
}

func (s *stream) Write(x interface{}) error {
	var f *DataFrame
	if s.f == nil {
		f = NewDataFrame(x, s.tag)
	} else {
		f = s.f.Copy(x, s.tag)
	}
	select {
	case s.down.input <- f:
		s.w++
		logrus.Debugf("%s write %s (%d)", s, f, s.w)
		return nil
	case <-s.ok:
		return io.EOF
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *stream) With(f *DataFrame) Stream {
	return &stream{
		id:  atomic.AddUint64(streamIds, 1),
		ctx: s.ctx,
		tag: s.tag,
		f:   f,

		input: s.input,
		ok:    s.ok,

		up:   s.up,
		down: s.down,
	}
}
