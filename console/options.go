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

func NewOptions(name string) *Options {
	return &Options{
		flag: flag.NewFlagSet(name, flag.ContinueOnError),
		args: make(map[int]*Option),
	}
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
			return errors.Wrapf(err, "arg %d", i)
		}
	}

	// Visit all of the flags and ensure that all have been set.
	// The flag library does not call set on options that have not been defined on the command line.
	p.flag.VisitAll(func(f *flag.Flag) {
		if err != nil {
			return
		}

		o, ok := f.Value.(*flagOption)
		if ok && !o.set {
			err = o.Set("")
		}
	})
	return err
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
