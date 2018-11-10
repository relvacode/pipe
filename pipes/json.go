package pipes

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/tap"
	"github.com/relvacode/pipe/valve"
	"io"
)

func init() {
	pipe.Pipes.Define(pipe.ModuleDefinition{
		Name: "json",
		Constructor: func(valve *valve.Control) pipe.Pipe {
			return JsonPipe{}
		},
	})
}

type JsonPipe struct {
}

func (JsonPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		v, err := stream.Read()
		if err != nil {
			return err
		}

		r, err := tap.Reader(v.Object)
		if err != nil {
			return err
		}

		decoder := json.NewDecoder(r)
		var i int
		for ; ; i++ {
			var x interface{}
			err := decoder.Decode(&x)
			if err != nil {
				tap.Close(r)
				if err == io.EOF {
					break
				}
				return err
			}
			err = stream.Write(x)
			if err != nil {
				tap.Close(r)
				return err
			}
		}
		if i == 0 {
			return errors.New("json: no data in stream")
		}
	}
}
