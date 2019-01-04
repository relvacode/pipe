package console

import (
	"flag"
	"fmt"
	"github.com/google/shlex"
	"github.com/pkg/errors"
	"strings"
)

type Usage interface {
	Usage() string
}

type flagOption struct {
	*Option
}

func (flagOption) String() string {
	return ""
}

func NewCommand() *Command {
	return &Command{
		flag: flag.NewFlagSet("", flag.ContinueOnError),
	}
}

// Command is given to a pipe to easily define input arguments to the pipe
type Command struct {
	o    *Option
	flag *flag.FlagSet
	args []*Option
}

// Usage returns the usage for this command.
func (c *Command) Usage() string {
	if c.o != nil {
		return c.o.Usage()
	}
	var args []string
	c.flag.VisitAll(func(f *flag.Flag) {
		args = append(args, fmt.Sprintf("-%s %s", f.Name, f.Value.(*flagOption).Usage()))
	})
	for _, o := range c.args {
		args = append(args, o.Usage())
	}

	return strings.Join(args, " ")
}

// Set parses and sets all of pointer values of described arguments
func (c *Command) Set(input string) error {
	if c.o != nil {
		return c.o.Set(input)
	}

	args, err := shlex.Split(input)
	if err != nil {
		return err
	}
	err = c.flag.Parse(args)
	if err != nil {
		return err
	}

	for i, a := range c.args {
		err = a.Set(c.flag.Arg(i))
		if err != nil {
			return err
		}
	}
	return err
}

func (c *Command) checkAnySet() {
	if c.o != nil {
		panic(errors.New("cannot call Any() more than once"))
	}
}

func (c *Command) Option(name string) *Option {
	c.checkAnySet()
	var o = &Option{
		name: name,
	}
	c.flag.Var(&flagOption{o}, name, "")
	return o
}

func (c *Command) Arg(n int) *Option {
	c.checkAnySet()
	if n != len(c.args) {
		panic(errors.Errorf("cannot call Arg(%d) out of order", n))
	}
	o := &Option{
		name: fmt.Sprint(n),
	}
	c.args = append(c.args, o)
	return o
}

// Any returns an Option that accepts any input (or none)
func (c *Command) Any() *Option {
	c.checkAnySet()
	c.o = new(Option)
	return c.o
}
