package console

import (
	"flag"
	"github.com/google/shlex"
	"github.com/pkg/errors"
)

type flagOption struct {
	*Option
}

func (flagOption) String() string {
	return ""
}

type Options struct {
	flag *flag.FlagSet
	args map[int]*Option
}

func (p *Options) Set(input string) error {
	args, err := shlex.Split(input)
	if err != nil {
		return err
	}
	err = p.flag.Parse(args)
	if err != nil {
		return err
	}

	for i, a := range p.args {
		err = a.Set(p.flag.Arg(i))
		if err != nil {
			return errors.Wrapf(err, "narg %d", i)
		}
	}
	return nil
}

func (p *Options) Option(name string) *Option {
	a := new(Option)
	p.flag.Var(&flagOption{a}, name, "")
	return a
}

func (p *Options) Arg(n int) *Option {
	a := new(Option)
	p.args[n] = a
	return a
}