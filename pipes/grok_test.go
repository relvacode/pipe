package pipes

import (
	"bytes"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/e2e"
	"github.com/relvacode/pipe/tap"
	"testing"
)

func TestGrokPipe(t *testing.T) {
	t.Run("ping", func(t *testing.T) {
		var (
			pattern = tap.Template(`icmp_seq=%{INT:seq} ttl=%{INT:ttl} time=%{NUMBER:rtt}`)
			pipes   = []pipe.Runnable{
				{Pipe: &GrokPipe{pattern: &pattern}},
			}
			inputs = []interface{}{
				bytes.NewReader([]byte(`64 bytes from 8.8.8.8: icmp_seq=1 ttl=116 time=9.540 ms`)),
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

		if r[`seq`] != "1" {
			t.Fatalf("Seq expected %v but got %v", "1", r[`seq`])
		}
		//if r[`ttl`] != 116 {
		//	t.Fatalf("Ttl expected %v but got %v", 116, r[`seq`])
		//}
		//if r[`rtt`] != 9.540 {
		//	t.Fatalf("Seq expected %v but got %v", 9.540, r[`seq`])
		//}
	})

}
