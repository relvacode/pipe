package pipes

import (
	"bytes"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/e2e"
	"io"
	"io/ioutil"
	"testing"
)

const testReplaceInput = `goodbye, world!`

func TestReplacePipe(t *testing.T) {
	var what = `goodbye`
	var with = `hello`
	var (
		pipes = []pipe.Runnable{
			{Pipe: &ReplacePipe{What: &what, With: &with}},
		}
		inputs = []interface{}{
			bytes.NewReader([]byte(testReplaceInput)),
		}
	)
	results, err := e2e.RunPipeTest(inputs, pipes)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected exactly 1 results")
	}

	s := results[0].Object.(io.Reader)
	b, err := ioutil.ReadAll(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello, world!" {
		t.Fatalf("expected %q but got %q", "hello, world!", string(b))
	}
}
