package aggregate

import (
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
)

func init() {
	pipe.Define(pipe.Pkg{
		Name: "sum",
		Constructor: func(command *console.Command) pipe.Pipe {
			return NewAggregator(command, func() Aggregation {
				return NewNumber(func(values []float64) (s float64) {
					for _, n := range values {
						s += n
					}
					return
				})
			})
		},
	})
}
