package pipe

import (
	"github.com/relvacode/pipe/console"
	"strings"
)

// A Builder constructs an instance of the module
// using the supplied arguments including how the module was called as the first argument.
type Builder func(args *console.Command) Pipe

// A Pkg describes a module in the Pipes.
type Pkg struct {
	Name        string
	Description string
	Constructor Builder
	Family      []Pkg
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
