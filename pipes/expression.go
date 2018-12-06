package pipes

import (
	"context"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
)

func init() {
	pipe.Define(
		Expr(
			"select",
			func(f *pipe.DataFrame, x interface{}, stream pipe.Stream) error {
				return stream.Write(nil, x)
			},
		))

	pipe.Define(
		Expr(
			"if",
			func(f *pipe.DataFrame, x interface{}, stream pipe.Stream) error {
				b, ok := x.(bool)
				if !ok {
					return errors.Errorf("expected boolean but expression returned %T", x)
				}
				if b {
					return stream.Write(nil, f.Object)
				}
				return nil
			},
		))
}

// ExprEvalFunc is a function called after calling an expression query.
// Given the query's return value, emit a value onto the stream or return error.
type ExprEvalFunc func(*pipe.DataFrame, interface{}, pipe.Stream) error

// Expr is a function that constructs a module definition for a JQ style query with
// additional logic applied to the return value.
func Expr(name string, f ExprEvalFunc) pipe.Pkg {
	return pipe.Pkg{
		Name: name,
		Constructor: func(console *console.Command) pipe.Pipe {
			return &ExprPipe{
				e: console.Any().Expression(),
				f: f,
			}
		},
	}
}

// ExprPipe executes a JQ style query and evaluates the result.
type ExprPipe struct {
	e *console.Expression
	f ExprEvalFunc
}

func (p *ExprPipe) Go(ctx context.Context, stream pipe.Stream) error {
	for {
		f, err := stream.Read(nil)
		if err != nil {
			return err
		}

		v, err := (*p.e).Eval(f.Context())
		if err != nil {
			return err
		}

		err = p.f(f, v, stream)
		if err != nil {
			return err
		}
	}
}
