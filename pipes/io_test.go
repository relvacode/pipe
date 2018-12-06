package pipes

import (
	"bytes"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/e2e"
	"io"
	"io/ioutil"
	"testing"
)

func TestBufferPipe(t *testing.T) {
	want := []byte("asdfghjkl")
	var b = bytes.NewBuffer(want)

	res, err := e2e.RunPipeTest([]interface{}{b}, []pipe.Runnable{
		{
			Pipe: BufferPipe{},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatalf("Expected length Of %d but got %d", 1, len(res))
	}

	cmp, err := ioutil.ReadAll(res[0].Object.(io.Reader))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(want, cmp) {
		t.Fatalf("Expecing %q but got %q", string(want), string(cmp))
	}

	_, err = b.Read(make([]byte, 1))
	if err != io.EOF {
		t.Fatalf("Expecting EOF but got %v", err)
	}
}
