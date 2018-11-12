package e2e

import "testing"

import (
	_ "github.com/relvacode/pipe/pipes"
)

type ConsoleTest struct {
	With   string
	Data   string
	Expect func(*testing.T, string, error)
}

func (dt ConsoleTest) Run(t *testing.T) {
	t.Run(dt.With, func(t *testing.T) {
		output, err := RunConsoleTest([]byte(dt.Data), dt.With)
		dt.Expect(t, output, err)
	})

}

type DSLTest struct {
	With   string
	Expect string
}

func (dt DSLTest) Run(t *testing.T) {
	c := ConsoleTest{
		With: dt.With,
		Data: `[{"a": 1, "b": "text", "c": {"a1": 1, "b1": 2}}]`,
		Expect: func(t *testing.T, result string, err error) {
			if err != nil {
				t.Fatal(err)
			}
			if result != dt.Expect {
				t.Fatalf("expected %q but got %q", dt.Expect, result)
			}
		},
	}
	c.Run(t)
}

func TestDSL(t *testing.T) {
	tests := []DSLTest{
		{
			With:   "json.decode::flatten::print {{this.b}}",
			Expect: "text",
		},
		{
			With:   "json.decode :: flatten :: print {{this.b}}",
			Expect: "text",
		},
		{
			With:   "json.decode :: flatten as o :: print {{o.b}}",
			Expect: "text",
		},
		{
			With:   "json.decode :: flatten as o::print {{o.b}}",
			Expect: "text",
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}

type InvalidParseTest struct {
	With string
}

func (it InvalidParseTest) Run(t *testing.T) {
	c := ConsoleTest{
		With: it.With,
		Expect: func(t *testing.T, result string, err error) {
			if err == nil {
				t.Fatalf("expected an error but got %q", result)
			}
		},
	}
	c.Run(t)
}

func TestDSLNoParse(t *testing.T) {
	tests := []InvalidParseTest{
		{
			With: "json as",
		},
		{
			With: "::",
		},
		{
			With: "::::",
		},
		{
			With: "",
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}
