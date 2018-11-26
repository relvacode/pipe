package pipe

import (
	"context"
	"fmt"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/tap"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
)

func init() {
	Define(ExecModule)
}

// StdinPipe emits a single os.Stdin object
type StdinPipe struct {
}

func (StdinPipe) Go(ctx context.Context, stream Stream) error {
	return stream.Write(os.Stdin)
}

// EchoPipe echos all output to stdout
type EchoPipe struct {
	Writer io.Writer
}

func (p *EchoPipe) Go(ctx context.Context, stream Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}

		switch o := f.Object.(type) {
		case io.Reader:
			io.Copy(p.Writer, o)
			tap.Close(o)
		default:
			fmt.Fprint(p.Writer, f.Object)
		}

		fmt.Print("\n")
	}
}

var ExecModule = Pkg{
	Name: "exec",
	Constructor: func(console *console.Command) Pipe {
		return &ExecPipe{
			command: console.Any().Template(),
		}
	},
}

// ExecPipe executes a command
type ExecPipe struct {
	command *tap.Template
}

func (p ExecPipe) execFrame(ctx context.Context, f *DataFrame, stream Stream) error {
	command, err := p.command.Render(f.Context())
	if err != nil {
		return err
	}

	logrus.Debugf("exec %q", command)

	// Use a custom IO pipe as the StdoutPipe closes the reader after Wait completes
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = pw

	err = cmd.Start()
	if err != nil {
		return err
	}
	defer pw.Close()

	err = stream.Write(pr)
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (p *ExecPipe) Go(ctx context.Context, stream Stream) error {
	for {
		f, err := stream.Read()
		if err != nil {
			return err
		}
		err = p.execFrame(ctx, f, stream)
		if err != nil {
			return err
		}
	}
}
