package main

import (
	"context"
	"flag"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/profile"
	"github.com/sirupsen/logrus"
	"os"
)

import (
	_ "github.com/relvacode/pipe/pipes"
	"github.com/relvacode/pipe/tap"
)

var (
	flagDebug = flag.Bool("debug", false, "debug mode")
	flagNoRc  = flag.Bool("norc", false, "Disable profile")
)

func Main() (func() pipe.RuntimeError, error) {
	flag.Parse()
	if *flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if !*flagNoRc {
		err := profile.Load()
		if err != nil {
			return nil, err
		}
	}

	defer func() {
		err := tap.Exit()
		if err != nil {
			logrus.Error(err)
		}
	}()

	r, err := pipe.ScriptReaderOf(flag.Arg(0))
	if err != nil {
		return nil, err
	}

	modules, err := pipe.Parse(r, pipe.Lib)
	if err != nil {
		return nil, err
	}
	tap.Close(r)

	var (
		ctx = context.Background()
		i   = &pipe.StdinPipe{}
		o   = &pipe.EchoPipe{
			Writer: os.Stdout,
		}
	)

	return func() pipe.RuntimeError {
		return pipe.RunIO(ctx, i, modules, o)
	}, nil
}

func main() {
	r, err := Main()
	if err != nil {
		logrus.Fatal(err)
	}

	errors := r()
	for _, err := range errors {
		logrus.Error(err)
	}
	if len(errors) > 0 {
		os.Exit(3)
	}
}
