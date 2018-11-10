package pipe

import "github.com/relvacode/pipe/valve"

// A Builder constructs an instance of the module
// using the supplied arguments including how the module was called as the first argument.
type Builder func(args *valve.Control) Pipe

// A ModuleDefinition describes a module in the Pipes.
type ModuleDefinition struct {
	Name        string
	Description string
	Constructor Builder
}

type Registry map[string]ModuleDefinition

func (r Registry) Define(m ModuleDefinition) {
	r[m.Name] = m
}

var (
	Pipes = make(Registry)
)
