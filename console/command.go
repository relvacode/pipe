package console

import "github.com/google/shlex"

func NewCommand(name string) *Command {
	return &Command{
		name:    name,
		Options: NewOptions(name),
	}
}

// Command is given to a pipe to easily define input arguments to the pipe
type Command struct {
	name  string
	apply apply
	*Options
}

func (c *Command) Name() string {
	return c.name
}

// Set parses and sets all of pointer values of described arguments
func (c *Command) Set(input string) error {
	if c.apply != nil {
		return c.apply(input)
	}
	return c.Options.Set(input)
}

// Split splits the command into basic strings using a shell-style parser
func (c *Command) Split() *[]string {
	var parts []string
	var ptr = &parts
	c.apply = func(input string) error {
		p, err := shlex.Split(input)
		if err != nil {
			return err
		}

		for _, s := range p {
			*ptr = append(*ptr, s)
		}
		return nil
	}
	return ptr
}

// Any returns an Option that accepts any input (or none)
func (c *Command) Any() *Option {
	var o = new(Option)
	c.apply = o.Set
	return o
}
