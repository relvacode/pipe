package pipes

import (
	"context"
	"encoding/csv"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"io"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "csv",
		Constructor: func(console *console.Command) pipe.Pipe {
			return CSVPipe{}
		},
	})
}

type CSVPipe struct {
}

func (CSVPipe) readStream(r io.Reader, stream pipe.Stream) error {
	var c = csv.NewReader(r)
	var headers []string

	for i := 0; ; i++ {
		record, err := c.Read()
		if err != nil {
			return err
		}

		if i == 0 {
			headers = record
			continue
		}

		row := make(map[string]string, len(headers))
		for i, k := range headers {
			row[k] = record[i]
		}
		err = stream.Write(nil, row)
		if err != nil {
			return err
		}
	}
}

func (p CSVPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		r, err := tap.Reader(f.Object)
		if err != nil {
			return err
		}

		err = p.readStream(r, stream)
		tap.Close(r)
		if err != nil {
			return err
		}
	}
}
