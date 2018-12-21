package pipe

import (
	"github.com/pkg/errors"
	"github.com/relvacode/pipe/console"
	"sort"
	"strings"
)

// A InitFn constructs an instance of the module
// using the supplied arguments including how the module was called as the first argument.
type InitFn func(*console.Command) Pipe

func Family(names ...string) string {
	if len(names) == 0 {
		panic(errors.New("need at least one name in package family"))
	}
	return strings.Join(names, ".")
}

// A Pkg describes a package - a pipe and/or a family of pipes.
type Pkg struct {
	// Name is the one-word name of this pipe or package.
	Name string
	// Constructor is a function to build an instance of this pipe.
	Constructor InitFn
}

type registry map[string]Pkg

// Sorted returns a sorted list of all install packages
func (r registry) Sorted() []Pkg {
	var keys = make([]string, len(r))
	var i int
	for k := range r {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var sorted = make([]Pkg, len(keys))
	for i, k := range keys {
		sorted[i] = r[k]
	}
	return sorted
}

var Lib = make(registry)

// Define registers the given package with the global library
func Define(pkg Pkg) {
	Lib[pkg.Name] = pkg
}
