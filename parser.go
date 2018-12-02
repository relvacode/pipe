package pipe

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe/console"
	"github.com/relvacode/pipe/dsl"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

// ScriptReaderOf obtains the script reader for the given command string.
// This string can either be a file name or a pipe script directly.
func ScriptReaderOf(command string) (io.Reader, error) {
	i, err := os.Stat(command)
	if os.IsNotExist(err) {
		return strings.NewReader(command), nil
	}
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	if i.IsDir() {
		return nil, errors.Errorf("%q is a directory not a script file", command)
	}

	return os.Open(command)
}

func Make(name string, cmd string, from registry) (Pipe, error) {
	logrus.Debugf("creating pipe %q using %q", name, cmd)
	if name == "" {
		return nil, errors.New("missing pipe name")
	}
	// Get the module from the Pipes
	c, ok := from[name]
	if !ok {
		c = ExecModule
		cmd = name + " " + cmd
	}

	var command = console.NewCommand(name)
	p := c.Constructor(command)
	err := command.Set(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "parse %q for %q", cmd, name)
	}
	return p, nil
}

func Parse(r io.Reader, reg registry) ([]Runnable, error) {
	pipes, err := dsl.Parse(r)
	if err != nil {
		return nil, err
	}

	var rn = make([]Runnable, len(pipes))
	for i, c := range pipes {
		p, err := Make(c.Name(), c.Args.String(), reg)
		if err != nil {
			return nil, errors.Wrapf(err, "create pipe %d", i)
		}
		rn[i] = Runnable{
			Tag:  NewTag(c.Tag()),
			Pipe: p,
		}
	}

	return rn, nil
}
