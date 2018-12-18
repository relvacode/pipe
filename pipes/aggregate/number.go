package aggregate

import (
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	"reflect"
	"strconv"
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

func NewNumber(reduce func([]float64) float64) *Number {
	return &Number{
		Reduce: reduce,
	}
}

// Number is an aggregator that collects float64 values and reduces them down into a single float64 value.
type Number struct {
	Values []float64
	Reduce func([]float64) float64
}

func (n *Number) Each(o interface{}) error {
	v := reflect.ValueOf(o)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n.Values = append(n.Values, float64(v.Int()))
	case reflect.Float32, reflect.Float64:
		n.Values = append(n.Values, v.Float())
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return err
		}
		n.Values = append(n.Values, f)
	default:
		return errors.Errorf("type %s is not a number", v.Kind())
	}

	return nil
}

func (n *Number) Final() (interface{}, error) {
	return n.Reduce(n.Values), nil
}
