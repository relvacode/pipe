package pipe

import (
	"context"
	"github.com/relvacode/pipe/console"
	"strings"
)

// A Builder constructs an instance of the module
// using the supplied arguments including how the module was called as the first argument.
type Builder func(*console.Command) Pipe

var _ Pipe = (Func)(nil)

// Func is a pipe implemented as a stateless function.
// Useful for creating simple pipes using FromFunc.
type Func func(context.Context, Stream) error

func (p Func) Go(ctx context.Context, stream Stream) error {
	return p(ctx, stream)
}

// Create an instance of a Pipe from a pipe function
func FromFunc(f Func) Builder {
	return func(*console.Command) Pipe {
		return f
	}
}

// A Pkg describes a package - a pipe and/or a family of pipes.
type Pkg struct {
	// Name is the one-word name of this pipe or package.
	Name string
	// Description is a brief one-line description on the purpose of the pipe.
	// Not used if a Constructor is not defined.
	Description string
	// Constructor is a function to build an instance of this pipe.
	Constructor Builder
	// Family is a list of additional sub-packages that belong to this package.
	// The final pipe name is the entire family tree joined with `.`
	Family []Pkg
}

type registry map[string]Pkg

var Lib = make(registry)

func define(m Pkg, family []string) {
	if m.Constructor != nil {
		Lib[strings.Join(family, ".")] = m
	}
	for _, f := range m.Family {
		define(f, append(family, strings.ToLower(f.Name)))
	}
}

func Define(m Pkg) {
	define(m, []string{strings.ToLower(m.Name)})
}
