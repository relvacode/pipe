package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/relvacode/pipe"
	"github.com/relvacode/pipe/profile"
	"github.com/sirupsen/logrus"
	"os"
)

import (
	_ "github.com/relvacode/pipe/pipes"
	"github.com/relvacode/pipe/tap"
)

var Rev string = "localbuild"

var (
	flagDebug = flag.Bool("debug", false, "Enable debug logging")
	flagNoRc  = flag.Bool("norc", false, "Disable profile")

	flagLibrary = flag.Bool("lib", false, "Get usage for all native modules then quit")
	flagPackage = flag.String("pkg", "", "Get usage for a specific package then quit")
)

func Main() error {
	if *flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugf("pipe %s", Rev)
	}

	if !*flagNoRc {
		err := profile.Load()
		if err != nil {
			return err
		}
	}

	if *flagPackage != "" {
		pkg, ok := pipe.Lib[*flagPackage]
		if !ok {
			return errors.Errorf("help: no such package %q", *flagPackage)
		}
		fmt.Println(pipe.Help(pkg))
		os.Exit(0)
	}

	if *flagLibrary {
		for _, pkg := range pipe.Lib.Sorted() {
			fmt.Println(pipe.Help(pkg))
		}
		os.Exit(0)
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

	modules, err := pipe.Parse(r, pipe.Lib)
	if err != nil {
		return err
	}
	_ = tap.Close(r)

	var (
		ctx = context.Background()
		i   = &pipe.StdinPipe{}
		o   = &pipe.EchoPipe{
			Writer: os.Stdout,
		}
	)

	return pipe.RunIO(ctx, i, modules, o).ErrorOrNil()
}

func main() {
	flag.Parse()

	err := Main()
	if err != nil {
		logrus.Fatal(err)
	}
}
