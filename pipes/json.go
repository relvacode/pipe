package pipes

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/tap"
	"io"
	"bytes"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "json",
		Family: []pipe.Pkg{
			{
				Name:        "decode",
				Constructor: pipe.FromFunc(JSONDecode),
			},
			{
				Name:        "encode",
				Constructor: pipe.FromFunc(JSONEncode),
			},
		},
	})
}

func JSONDecode(_ context.Context, stream pipe.Stream) error {
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

func JSONEncode(_ context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		var buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(f.Object)
		if err != nil {
			return err
		}

		err = stream.Write(buf)
		if err != nil {
			return err
		}
	}
}
