package dsl

import (
	"github.com/SteelSeries/bufrr"
	"github.com/pkg/errors"
	"io"
)

var (
	// EOP is an end of pipe error
	EOP = errors.New("End of pipe")
)

// IsDoubleRune returns true if the current position of the RunePeeker is Args double instance of rune r.
// The characters in the RunePeeker are consumed if matched.
func IsDoubleRune(b bufrr.RunePeeker, x, y rune) bool {
	p, _, _ := b.PeekRune()
	if p != x {
		return false
	}
	b.ReadRune()
	p, _, _ = b.PeekRune()
	if p != y {
		b.UnreadRune()
		return false
	}

	b.ReadRune()
	return true
}

func IsNextPipe(b bufrr.RunePeeker) bool {
	return IsDoubleRune(b, ':', ':')
}

func IsStartTag(b bufrr.RunePeeker) bool {
	return IsDoubleRune(b, 'a', 's')
}

func IsEOF(b bufrr.RunePeeker) bool {
	r, _, err := b.PeekRune()
	return err == io.EOF || r == bufrr.EOF
}
