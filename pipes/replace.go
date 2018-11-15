package pipes

import (
	"bufio"
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/relvacode/rewrite"
	"io"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "replace",
		Constructor: func(console *console.Command) pipe.Pipe {
			args := console.SplitNArgs(2)
			return &ReplacePipe{
				What: args[0],
				With: args[1],
			}
		},
	})
}

func NewStreamReplacer(r io.Reader, what, with []byte) *StreamReplacer {
	return &StreamReplacer{
		From:    bufio.NewReader(r),
		What:    what,
		With:    with,
		replace: new(bytes.Buffer),
	}
}

// StreamReplacer is a content replacing proxy for a byte stream
type StreamReplacer struct {
	From *bufio.Reader
	What []byte
	With []byte

	doReplace bool
	replace   *bytes.Buffer
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (sr *StreamReplacer) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	var n int
	if sr.doReplace {
		for {
			rn, err := sr.replace.Read(p[n:])
			n += rn

			if err == io.EOF {
				break
				sr.doReplace = false
			}

			if n == len(p) {
				return n, nil
			}
		}

	}

	// If the next n bytes are exactly the match condition
	peek, err := sr.From.Peek(len(sr.What))
	if err != nil && err != io.EOF {
		return n, err
	}

	// Read until a the first known character
	idx := bytes.IndexByte(peek, sr.What[0])

	// Read position is exactly at the text we want to replace
	if idx == 0 && bytes.Equal(sr.What, peek) {
		sr.replace.Reset()
		sr.From.Discard(len(sr.What))

		// Attempt to copy what we can inside this call
		diff := len(p) - n
		if diff > 0 {
			rn := copy(p[n:], sr.With)
			n += rn
		}

		// Diff happens to be exactly or more the length of the replacement string
		if diff >= len(sr.With) {
			rn, err := sr.Read(p[n:])
			return n + rn, err
		}

		// Write the remaining data later
		sr.replace.Write(sr.With[diff:])
		sr.doReplace = true
		rn, err := sr.Read(p[n:])
		return n + rn, err
	}

	if idx > 0 {
		// Read until the correct index
		rn, err := sr.From.Read(p[n:min(len(p), idx)])
		return n + rn, err
	}

	// Nothing found
	rn, err := sr.From.Read(p[n:min(len(p), len(sr.What))])
	return n + rn, err
}

func (sr *StreamReplacer) Close() error {
	return tap.Close(sr.From)
}

type ReplacePipe struct {
	What *string
	With *string
}

func (p *ReplacePipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		wf, err := f.Var(*p.What)
		if err != nil {
			return err
		}
		if wf == "" {
			return errors.New("cannot search for empty string in replace")
		}

		wd, err := f.Var(*p.With)
		if err != nil {
			return err
		}

		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		replacer := rewrite.New(r, []byte(wf), []byte(wd))
		err = stream.Write(replacer)
		if err != nil {
			return err
		}
	}
}
