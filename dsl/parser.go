package dsl

import (
	"github.com/SteelSeries/bufrr"
	"io"
)

func Parse(r io.Reader) ([]*Command, error) {
	var p = new(Pipe)
	var b = bufrr.NewReader(r)
	var err = p.Read(b)
	return p.pipes, err
}
