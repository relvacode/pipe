package pipes

import (
	"bytes"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/e2e"
	"testing"
)

const testSplitInput = `a
b
c
d`

func TestSplitPipe(t *testing.T) {
	var split = "\n"
	var (
		pipes = []pipe.Runnable{
			{Pipe: &SplitPipe{Split: &split}},
		}
		inputs = []interface{}{
			bytes.NewReader([]byte(testSplitInput)),
		}
	)
	results, err := e2e.RunPipeTest(inputs, pipes)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 4 {
		t.Fatalf("Expected exactly 4 results")
	}

	s := results[0].Object.(string)
	if s != "a" {
		t.Fatalf("result 0 expected %q but got %q", "a", s)
	}
}
