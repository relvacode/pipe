package main

import (
	"context"
	"flag"
	"fmt"
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

func Main() error {
	flag.Parse()
	if *flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if !*flagNoRc {
		err := profile.Load()
		if err != nil {
			return err
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
		return err
	}

	modules, err := pipe.Parse(r, pipe.Pipes)
	if err != nil {
		return err
	}
	tap.Close(r)

	var (
		ctx = context.Background()
		i   = &pipe.StdinPipe{}
		o   = &pipe.EchoPipe{
			Writer: os.Stdout,
		}
	)

	return pipe.RunIO(ctx, i, modules, o)
}

func main() {
	err := Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR ", err)
		os.Exit(1)
	}
}
