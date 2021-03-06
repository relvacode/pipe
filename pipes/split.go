package pipes

import (
	"bufio"
	"context"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
	"strings"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "split",
		Constructor: func(console *console.Command) pipe.Pipe {
			return &SplitPipe{
				Split: console.Arg(0).Default("\n").String(),
			}
		},
	})
}

func SplitAt(substring string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {

		// Return nothing if at end Of file and no data passed
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Find the index Of the input Of the separator substring
		if i := strings.Index(string(data), substring); i >= 0 {
			return i + len(substring), data[0:i], nil
		}

		// If at end Of file with data return the data
		if atEOF {
			return len(data), data, nil
		}

		return
	}
}

// Split splits an input reader or string into lines
type SplitPipe struct {
	Split *string
}

func (p *SplitPipe) splitStream(r io.Reader, stream pipe.Stream) error {
	s := bufio.NewScanner(r)
	s.Split(SplitAt(*p.Split))
	for s.Scan() {
		err := stream.Write(nil, s.Text())
		if err != nil {
			return err
		}
	}
	return s.Err()
}

func (p *SplitPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		v, err := stream.Read(nil)
		if err != nil {
			return err
		}

		r, err := tap.Reader(v.Object)
		if err != nil {
			return err
		}

		err = p.splitStream(r, stream)
		tap.Close(r)
		if err != nil {
			return err
		}
	}
}
