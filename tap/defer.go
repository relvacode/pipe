package tap

import (
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"sync"
)

// Deferred is a function to be called when the program exits.
// This should be used to clean up anything created by pipes in this program.
type Deferred func() error

type exitHandler struct {
	mtx sync.Mutex
	f   []Deferred
}

var onexit *exitHandler

func init() {
	onexit = new(exitHandler)
}

// Defer a function to be called before the program exits
func Defer(f Deferred) {
	onexit.mtx.Lock()
	onexit.f = append(onexit.f, f)
	onexit.mtx.Unlock()
}

// Exit should be called just before the program exits
func Exit() (err *multierror.Error) {
	onexit.mtx.Lock()
	defer onexit.mtx.Unlock()

	for _, f := range onexit.f {
		logrus.Debug("execute deferred function %v", f)
		err = multierror.Append(err, f())
	}
	return err
}
