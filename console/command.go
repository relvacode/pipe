package console

import (
	"flag"
)

func NewCommand(name string) *Command {
	return &Command{
		name:   name,
		Option: new(Option),
	}
}

// Command is given to a pipe to easily define input arguments to the pipe
type Command struct {
	name    string
	options *Options
	*Option
}

func (c *Command) Name() string {
	return c.name
}

// Set parses and sets all of pointer values of described arguments
func (c *Command) Set(input string) error {
	if c.options != nil {
		return c.options.Set(input)
	}
	if c.Option.apply != nil {
		return c.Option.Set(input)
	}
	return nil
}

func (c *Command) Options() *Options {
	c.options = &Options{
		flag: flag.NewFlagSet(c.name, flag.ContinueOnError),
		args: make(map[int]*Option),
	}
	return c.options
}
