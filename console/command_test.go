package console

import "testing"

func TestOptions_Set(t *testing.T) {
	w := "123"
	o := NewCommand()
	a := o.Option("ab").Default(w).String()
	err := o.Set("")
	if err != nil {
		t.Fatal(err)
	}

	if w != *a {
		t.Fatalf("Wanted %q; got %q", w, *a)
	}
}
