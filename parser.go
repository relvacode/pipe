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
	forkStart      = '('
	forkEnd        = ')'
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
	cmd = strings.TrimSpace(cmd)
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

	var command = console.NewCommand(name)
	p := c.Constructor(command)
	err := command.Set(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "parse %q for %q", cmd, name)
	}
	return p, nil
}

func newPipeScanner(r *bufrr.Reader, registry registry) *pipeScanner {
	return &pipeScanner{
		r:   r,
		reg: registry,
	}
}

type pipeScanner struct {
	reader io.Reader
	r      *bufrr.Reader

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
	if !(r == delimiter || r == forkEnd) {
		return false
	}
	nr, _, _ := s.r.PeekRune()
	if nr == eof {
		return true
	}
	if nr == delimiter || r == forkEnd {
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

func (s *pipeScanner) scanTag() error {
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

func (s *pipeScanner) Scan() (bool, error) {
	var isAtEndPipe bool

name:
	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return false, err
		}
		if r == eof {
			// EOF at command name but we do have a name
			if s.nb.Len() > 0 {
				return false, nil
			}
			return false, io.EOF
		}

		switch {
		case r == forkStart && s.nb.Len() == 0:
			nr, _, _ := s.r.PeekRune()
			if nr == forkStart {
				s.r.ReadRune()
				return true, nil
			}
		case r == comment:
			err = s.readComment()
			if err != nil {
				return false, err
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
			return false, errors.Errorf("unexpected character %q", string(r))
		}
	}

	if isAtEndPipe {
		return false, nil
	}

args:
	for {
		r, _, err := s.r.ReadRune()
		if err != nil {
			return false, err
		}

		switch {
		case r == eof:
			return false, nil
		case r == comment:
			err = s.readComment()
			if err != nil {
				return false, err
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
			//
			//// No command yet, empty spaces
			//if s.cb.Len() == 0 {
			//	continue args
			//}
		}

		s.cb.WriteRune(r)
	}

	if isAtEndPipe {
		return false, nil
	}

	return false, s.scanTag()
}

func (s *pipeScanner) Next() (*Runnable, error) {
	defer s.Reset()

	fork, err := s.Scan()
	if err != nil {
		return nil, err
	}

	if fork {
		logrus.Debugf("start parsing forked pipe")
		forked, err := parse(s.r, s.reg)
		if err != nil {
			return nil, err
		}

		return &Runnable{
			Pipe: ForkPipe(forked),
		}, nil
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
	return parse(bufrr.NewReader(r), reg)
}

func parse(r *bufrr.Reader, reg registry) ([]Runnable, error) {
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
