package e2e

import (
	"bytes"
	"context"
	"github.com/relvacode/pipe"
	"strings"
)

type WTestPipe struct {
	Objects []interface{}
}

func (p *WTestPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for i := 0; i < len(p.Objects); i++ {
		err := stream.Write(p.Objects[i])
		if err != nil {
			return err
		}
	}
	return nil
}

type RTestPipe struct {
	Results []*pipe.DataFrame
}

func (p *RTestPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}
		p.Results = append(p.Results, f)
	}
}

func RunPipeTest(inputs []interface{}, pipes []pipe.Runnable) ([]*pipe.DataFrame, error) {
	exec := make([]pipe.Runnable, len(pipes)+2)
	copy(exec[1:], pipes)
	exec[0] = pipe.Runnable{
		Pipe: &WTestPipe{
			Objects: inputs,
		},
	}
	r := new(RTestPipe)
	exec[len(exec)-1] = pipe.Runnable{
		Pipe: r,
	}

	err := pipe.Run(context.Background(), exec)
	return r.Results, err
}

func RunConsoleTest(stdin []byte, command string) (string, error) {
	pipes, err := pipe.Parse(strings.NewReader(command), pipe.Lib)
	if err != nil {
		return "", err
	}

	exec := make([]pipe.Runnable, len(pipes)+2)
	copy(exec[1:], pipes)
	exec[0] = pipe.Runnable{
		Pipe: &WTestPipe{
			Objects: []interface{}{
				bytes.NewReader(stdin),
			},
		},
	}

	var b = new(bytes.Buffer)
	r := &pipe.EchoPipe{
		Writer: b,
	}
	exec[len(exec)-1] = pipe.Runnable{
		Pipe: r,
	}
	err = pipe.Run(context.Background(), exec)
	return b.String(), err
}
