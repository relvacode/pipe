package console

import (
	"testing"
)

func TestOptionDefault(t *testing.T) {
	t.Run("no default", func(t *testing.T) {
		var a = new(Option)
		a.String()
		err := a.Set("")
		if err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("string", func(t *testing.T) {
		var s = "s"
		var a = new(Option).Default(s)
		var x = a.String()
		err := a.Set("")
		if err != nil {
			t.Fatal(err)
		}
		if *x != s {
			t.Fatalf("Expected %q but got %q", s, *x)
		}
	})
}
