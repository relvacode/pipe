package pipe

import (
	"context"
	"github.com/relvacode/pipe/console"
	"strings"
	"testing"
)

type TestPipe struct{}

func (TestPipe) Go(context.Context, Stream) error {
	return nil
}

func TestParse(t *testing.T) {
	t.Run("pipeline 1", func(t *testing.T) {
		i, err := Parse(strings.NewReader("test"), registry{
			"test": Pkg{
				Constructor: func(*console.Command) Pipe {
					return TestPipe{}
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(i) != 1 {
			t.Fatalf("Expected 1 parsed modules, got %d", len(i))
		}
		_, ok := i[0].Pipe.(TestPipe)
		if !ok {
			t.Fatalf("expected pipe instance to be %T not %T", TestPipe{}, i[0].Pipe)
		}
	})

	t.Run("exec fallback", func(t *testing.T) {
		i, err := Parse(strings.NewReader("exec a b c"), registry{})
		if err != nil {
			t.Fatal(err)
		}
		if len(i) != 1 {
			t.Fatalf("Expected 1 parsed modules, got %d", len(i))
		}
		e, ok := i[0].Pipe.(*ExecPipe)
		if !ok {
			t.Fatalf("expected pipe instance to be %T not %T", new(ExecPipe), i[0].Pipe)
		}
		var want = "a b c"
		if string(*e.args) != want {
			t.Fatalf("expected args %s but got %s", want, *e.args)
		}
	})
}
