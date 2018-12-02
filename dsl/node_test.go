package dsl

import (
	"github.com/SteelSeries/bufrr"
	"strings"
	"testing"
)

type ParserTestCase struct {
	Using  string
	With   Node
	Expect string
}

func (tc ParserTestCase) Run(t *testing.T) {
	r := bufrr.NewReader(strings.NewReader(tc.Using))
	err := tc.With.Read(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", tc.With)
	got := tc.With.String()
	if got != tc.Expect {
		t.Fatalf("Wanted %q but got %q", tc.Expect, got)
	}
}

func TestPipe_Read(t *testing.T) {
	cases := []ParserTestCase{
		{
			Using:  "Args one two three as foo::b four five six as bar",
			With:   new(Pipe),
			Expect: "Args one two three as foo :: b four five six as bar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Using, tc.Run)
	}
}
