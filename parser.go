package pipe

import (
	"bytes"
	"github.com/SteelSeries/bufrr"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe/console"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"unicode"
)

const (
	delimiter      = ':'
	comment        = '#'
	eof       rune = bufrr.EOF
)

// ScriptReaderOf obtains the script reader for the given command string.
// This string can either be a file name or a pipe script directly.
func ScriptReaderOf(command string) (io.Reader, error) {
	i, err := os.Stat(command)
	if os.IsNotExist(err) {
		return strings.NewReader(command), nil
	}
	if err != nil {
		return nil, err
	}
	if i.IsDir() {
		return nil, errors.Errorf("%q is a directory not a script file", command)
	}

	return os.Open(command)
}

func MakePipe(name string, cmd string, reg registry) (Pipe, error) {
	logrus.Debugf("creating pipe %q using %q", name, cmd)

	if name == "" {
		return nil, errors.New("missing command")
	}
	// Get the module from the Pipes
	c, ok := reg[name]
	if !ok {
		c = ExecModule
		cmd = name + " " + cmd
	}

	var control = new(console.Command)
	p := c.Constructor(control)
	err := control.Parse(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "parse %q for %q", cmd, name)
	}
	return p, nil
}

func newPipeScanner(r io.Reader, registry registry) *pipeScanner {
	return &pipeScanner{
		r:   bufrr.NewReader(r),
		reg: registry,
	}
}

type pipeScanner struct {
	r *bufrr.Reader

	nb bytes.Buffer
	cb bytes.Buffer
	tb bytes.Buffer

	reg registry
	len int // counts number of parsed modules (at position)
}

func (s *pipeScanner) Reset() {
	s.nb.Reset()
	s.cb.Reset()
	s.tb.Reset()
}

func (s *pipeScanner) isAtEndPipe(r rune) bool {
	if r == eof {
		return true
	}
	if r != delimiter {
		return false
	}
	nr, _, _ := s.r.PeekRune()
	if nr == eof {
		return true
	}
	if nr == delimiter {
		s.r.ReadRune()
		return true
	}
	return false
}

func (s *pipeScanner) readComment() error {
	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return err
		}
		switch r {
		case '\n':
			return nil
		case '\r':
			nr, _, err := s.r.PeekRune()
			if err != nil {
				return err
			}
			if nr == '\n' {
				_, _, err = s.r.ReadRune()
				return err
			}
			return nil
		}
	}
}

func (s *pipeScanner) Scan() error {
	var isAtEndPipe bool

name:
	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return err
		}
		if r == eof {
			// EOF at command name but we do have a name
			if s.nb.Len() > 0 {
				return nil
			}
			return io.EOF
		}

		switch {
		case r == comment:
			err = s.readComment()
			if err != nil {
				return err
			}
		case s.isAtEndPipe(r):
			isAtEndPipe = true
			break name
		case unicode.IsSpace(r):
			if s.nb.Len() > 0 {
				s.r.UnreadRune()
				break name
			}
			// Strip leading spaces
			continue
		case unicode.IsDigit(r) || unicode.IsLetter(r) || r == '.' || r == '/':
			s.nb.WriteRune(r)
		default:
			return errors.Errorf("unexpected character %q", string(r))
		}
	}

	if isAtEndPipe {
		return nil
	}

args:
	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return err
		}

		switch {
		case r == eof:
			return nil
		case r == comment:
			err = s.readComment()
			if err != nil {
				return err
			}
		case s.isAtEndPipe(r):
			isAtEndPipe = true
			break args
		case unicode.IsSpace(r):
			nr, _, _ := s.r.PeekRune()
			if nr == 'a' {
				s.r.ReadRune()
				enr, _, _ := s.r.PeekRune()
				if enr == 's' {
					s.r.ReadRune()
					break args
				}
				s.r.UnreadRune()
			}

			// No command yet, empty spaces
			if s.cb.Len() == 0 {
				continue args
			}
		}

		s.cb.WriteRune(r)
	}

	if isAtEndPipe {
		return nil
	}

	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return err
		}
		switch {
		case r == eof:
			if s.tb.Len() == 0 {
				return errors.New("expected tag name after `as`")
			}

			return nil
		case r == comment:
			err = s.readComment()
			if err != nil {
				return err
			}
		case unicode.IsSpace(r):
			continue
		case s.isAtEndPipe(r):
			return nil
		}
		s.tb.WriteRune(r)
	}
}

func (s *pipeScanner) Next() (*Runnable, error) {
	defer s.Reset()

	err := s.Scan()
	if err != nil {
		return nil, err
	}

	m, err := MakePipe(s.nb.String(), s.cb.String(), s.reg)
	if err != nil {
		return nil, err
	}

	s.len++

	return &Runnable{
		Tag:  NewTag(s.tb.String()),
		Pipe: m,
	}, nil
}

func Parse(r io.Reader, reg registry) ([]Runnable, error) {
	var (
		modules []Runnable
		scanner = newPipeScanner(r, reg)
	)
	for {
		m, err := scanner.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrapf(err, "pipe %d", scanner.len)
		}

		logrus.Debugf("found runnable %s at position %d", m, scanner.len-1)
		modules = append(modules, *m)
	}

	if len(modules) == 0 {
		return nil, errors.Errorf("no pipes described")
	}

	return modules, nil
}
