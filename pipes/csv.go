package pipes

import (
	"context"
	"encoding/csv"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/tap"
	"github.com/relvacode/pipe/valve"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "csv",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return CsvPipe{}
		},
	})
}

type CsvPipe struct {
}

func (CsvPipe) streamReader(c *csv.Reader, stream pipe.Stream) error {
	for {
		record, err := c.Read()
		if err != nil {
			return err
		}
		err = stream.Write(record)
		if err != nil {
			return err
		}
	}
}

func (p CsvPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		v, err := stream.Read()
		if err != nil {
			return err
		}

		r, err := tap.Reader(v)
		if err != nil {
			return err
		}

		c := csv.NewReader(r)
		err = p.streamReader(c, stream)
		tap.Close(r)
		if err != nil {
			return err
		}
	}
}
