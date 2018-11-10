// package tap provides common interfaces for data exchange
package tap

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

// Reader gets a reader interface from an input interface
func Reader(x interface{}) (io.Reader, error) {
	switch v := x.(type) {
	case io.Reader:
		return v, nil
	case string:
		return strings.NewReader(v), nil
	}
	return nil, errors.Errorf("Expected a string or file-like object but got type %T", x)
}

type closer interface {
	Close() error
}

func Close(x interface{}) {
	c, ok := x.(closer)
	if ok {
		logrus.Debugf("close %T", x)
		c.Close()
	}
}
