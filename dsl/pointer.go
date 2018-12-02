package dsl

import (
	"github.com/SteelSeries/bufrr"
	"github.com/pkg/errors"
)

var _ bufrr.RunePeeker = (*RuneSeekPointer)(nil)

type RuneSeekPointer struct {
	bufrr.RunePeeker

	char int
}

// Pos returns the current position in the reader
func (rsp *RuneSeekPointer) Pos() int {
	return rsp.char
}

func (rsp *RuneSeekPointer) Err(err error) error {
	r, _, _ := rsp.RunePeeker.PeekRune()
	return errors.Wrapf(err, "at character %d (%q)", rsp.Pos(), r)
}

func (rsp *RuneSeekPointer) UnreadRune() (err error) {
	err = rsp.RunePeeker.UnreadRune()
	if err == nil {
		rsp.char--
	}

	return
}

func (rsp *RuneSeekPointer) ReadRune() (r rune, size int, err error) {
	r, size, err = rsp.RunePeeker.ReadRune()
	if err == nil {
		rsp.char++
	}

	return
}
