package e2e

import (
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/console"
	_ "github.com/relvacode/pipe/pipes"
	"testing"
)

func TestLibrary(t *testing.T) {
	for _, pkg := range pipe.Lib {
		t.Run(pkg.Name, func(t *testing.T) {
			cmd := console.NewCommand()

			t.Run("can construct", func(t *testing.T) {
				if pkg.Constructor(cmd) == nil {
					t.Fatal("produced a nil pipe")
				}
			})
			t.Run("can usage", func(t *testing.T) {
				cmd.Usage()
			})
		})
	}
}
