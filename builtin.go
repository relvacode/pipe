package pipe

import (
	"context"
	"fmt"
	"github.com/google/shlex"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
)

// StdinPipe emits a single os.Stdin object
type StdinPipe struct {
}

func (StdinPipe) Go(ctx context.Context, stream Stream) error {
	return stream.Write(nil, os.Stdin)
}

// EchoPipe echos all output to stdout
type EchoPipe struct {
	Writer io.Writer
}

func (p *EchoPipe) Go(ctx context.Context, stream Stream) error {
	for i := 0; ; i++ {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		if i > 0 {
			fmt.Print("\n")
		}

		switch o := f.Object.(type) {
		case io.Reader:
			io.Copy(p.Writer, o)
			tap.Close(o)
		default:
			fmt.Fprint(p.Writer, f.Object)
		}

	}
}

func NewExecPkg(name string) Pkg {
	return Pkg{
		Name: name,
		Constructor: func(console *console.Command) Pipe {
			return &ExecPipe{
				name: name,
				args: console.Any().Default("").Template(),
			}
		},
	}
}

// ExecPipe executes a args
type ExecPipe struct {
	name string
	args *tap.Template
}

func (p ExecPipe) execFrame(ctx context.Context, f *DataFrame, stream Stream) error {
	fa, err := p.args.Render(f.Context())
	if err != nil {
		return err
	}

	args, err := shlex.Split(fa)
	if err != nil {
		return err
	}

	logrus.Debugf("exec %q", args)

	// Use a custom IO pipe as the StdoutPipe closes the reader after Wait completes
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	r, _ := tap.Reader(f.Object)

	cmd := exec.CommandContext(ctx, p.name, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = r
	cmd.Stderr = os.Stderr
	cmd.Stdout = pw

	err = cmd.Start()
	if err != nil {
		return err
	}
	defer pw.Close()

	err = stream.Write(nil, pr)
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (p *ExecPipe) Go(ctx context.Context, stream Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}
		err = p.execFrame(ctx, f, stream)
		if err != nil {
			return err
		}
	}
}
