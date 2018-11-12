package console

import (
	"github.com/antonmedv/expr"
	"github.com/google/shlex"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type apply func(string) error

// Command is given to a pipe to easily define input arguments to the pipe
type Command struct {
	apply apply
}

// Parse sets all of pointer values of described arguments
func (c *Command) Parse(input string) error {
	if c.apply == nil {
		return nil
	}
	return c.apply(strings.TrimSpace(input))
}

// Input returns all input as one
func (c *Command) Input() *Arg {
	var a = new(Arg)
	c.apply = func(s string) error {
		return a.final(s)
	}
	return a
}

// Args returns n amount of required arguments
func (c *Command) Args(n int) []*Arg {
	args := make([]*Arg, n)
	for i := range args {
		args[i] = new(Arg)
	}
	c.apply = func(s string) error {
		parsed, err := shlex.Split(s)
		if err != nil {
			return err
		}
		if len(parsed) != n {
			return errors.Errorf("expected %d arguments (given %d)", n, len(parsed))
		}

		for i, a := range args {
			err = a.final(parsed[i])
			if err != nil {
				return errors.Wrapf(err, "arg %d (%q)", i, parsed[i])
			}
		}

		return nil
	}

	return args
}

// Arg describes the conversion of an input string into a destination pointer value
type Arg struct {
	final func(s string) error
}

// String is a required string
func (a *Arg) String() *string {
	var value string
	var ptr = &value
	a.final = func(s string) error {
		if s == "" {
			return errors.New("no argument provided")
		}
		*ptr = s
		return nil
	}
	return ptr
}

// DefaultString is an optional string
func (a *Arg) DefaultString(d string) *string {
	var value string
	var ptr = &value
	a.final = func(s string) error {
		if s == "" {
			*ptr = d
		} else {
			*ptr = s
		}
		return nil
	}
	return ptr
}

// Int parses as an 64 bit integer
func (a *Arg) Int() *int64 {
	var value int64
	var ptr = &value
	a.final = func(s string) (error) {
		var err error
		*ptr, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		return nil
	}
	return ptr
}

// Expression parses an expr expression
func (a *Arg) Expression() *expr.Node {
	var n expr.Node
	var ptr = &n
	a.final = func(s string) error {
		pn, err := expr.Parse(s)
		if err != nil {
			return err
		}
		*ptr = pn
		return nil
	}
	return ptr
}

// Duration parses a duration
func (a *Arg) Duration() *time.Duration {
	var d time.Duration
	var ptr = &d
	a.final = func(s string) error {
		pd, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*ptr = pd
		return nil
	}
	return ptr
}

// Args returns all arguments split using a shell-like parser
func (a *Arg) Args() *[]string {
	var args = make([]string, 0)
	var ptr = &args
	a.final = func(s string) error {
		parsed, err := shlex.Split(s)
		if err != nil {
			return err
		}
		for _, p := range parsed {
			*ptr = append(*ptr, p)
		}
		return nil
	}
	return ptr
}
