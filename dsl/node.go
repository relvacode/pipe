package dsl

import (
	"bytes"
	"fmt"
	"github.com/SteelSeries/bufrr"
	"io"
	"strings"
	"unicode"
)

type Node interface {
	String() string
	// Read reads from b until the token is fully read
	Read(b bufrr.RunePeeker) error
}

type Tag struct {
	b bytes.Buffer
}

func (t *Tag) String() string {
	if t.b.Len() == 0 {
		return ""
	}
	return fmt.Sprintf("as %s", t.b.String())
}

func (t *Tag) Read(b bufrr.RunePeeker) error {
	for {
		r, _, err := b.PeekRune()
		if err != nil {
			return err
		}

		switch {
		case IsEOF(b):
			if t.b.Len() == 0 {
				return io.ErrUnexpectedEOF
			}
			return io.EOF
		case IsNextPipe(b):
			return EOP
		case unicode.IsSpace(r):
		default:
			t.b.WriteRune(r)
		}

		b.ReadRune()
	}
}

type Arg struct {
	b bytes.Buffer
	t Tag
}

func (a *Arg) String() string {
	return a.b.String()
}

func (a *Arg) Read(b bufrr.RunePeeker) error {
	for {
		r, _, err := b.PeekRune()
		if err != nil {
			return err
		}

		switch {
		case IsEOF(b):
			return io.EOF
		case IsNextPipe(b):
			return EOP
		case IsStartTag(b):
			return a.t.Read(b)
		case a.b.Len() == 0 && unicode.IsSpace(r):
		default:
			a.b.WriteRune(r)
		}

		b.ReadRune()
	}
}

type Command struct {
	b    bytes.Buffer
	Args Arg
}

func (c *Command) Name() string {
	return c.b.String()
}

func (c *Command) Tag() string {
	return c.Args.t.b.String()
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s%s", c.b.String(), c.Args.String(), c.Args.t.String())
}

func (c *Command) Read(b bufrr.RunePeeker) error {
	for {
		r, _, err := b.PeekRune()
		if err != nil {
			return err
		}

		switch {
		case IsEOF(b):
			if c.b.Len() == 0 {
				return io.ErrUnexpectedEOF
			}
			return io.EOF
		case IsNextPipe(b):
			return EOP
		case c.b.Len() == 0 && unicode.IsSpace(r):
		case c.b.Len() > 0 && unicode.IsSpace(r):
			b.ReadRune()
			return c.Args.Read(b)
		default:
			c.b.WriteRune(r)
		}

		b.ReadRune()
	}
}

type Pipe struct {
	pipes []*Command
}

func (p *Pipe) String() string {
	var s strings.Builder
	for i, c := range p.pipes {
		s.WriteString(c.String())
		if i < len(p.pipes)-1 {
			s.WriteString(" :: ")
		}
	}
	return s.String()
}

func (p *Pipe) Read(b bufrr.RunePeeker) error {
	r := &RuneSeekPointer{
		RunePeeker: b,
	}
	for {
		var c = new(Command)
		err := c.Read(b)
		switch err {
		case io.EOF:
			p.pipes = append(p.pipes, c)
			return nil
		case EOP:
			p.pipes = append(p.pipes, c)
		default:
			return r.Err(err)
		}
	}
}
