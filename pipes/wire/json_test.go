package wire

import (
	"bytes"
	"fmt"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/e2e"
	"testing"
)

func TestJsonPipe(t *testing.T) {
	var (
		pipes = []pipe.Runnable{
			{Pipe: &Pipe{Protocol: JSONProtocol{}}},
		}
		inputs = []interface{}{
			bytes.NewReader([]byte(`{"a": 1, "b": 2}`)),
		}
	)
	results, err := e2e.RunPipeTest(inputs, pipes)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected exactly one result")
	}

	r, ok := results[0].Object.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} but got %T", results[0].Object)
	}

	a := r["a"]
	if a.(float64) != 1 {
		t.Fatalf("Expected 1 got %v", a)
	}
}
