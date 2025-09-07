package main

import (
	"context"
	"flag"
	"fmt"
	"oci-runtime/internal/app"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/platform/config"
	"oci-runtime/internal/platform/logging"
	"os"
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `oci-runtime
Usage:
  oci-runtime run
`)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}
	subcmd := flag.Arg(0)

	_ = config.Load()
	logger := logging.New(subcmd)
	ctx := context.Background()

	runHandler := mw.Chain(
		app.NewRunHandler(),
		mw.WithLogging[app.RunCmd]("app", logger),
	)

	initHandler := mw.Chain(
		app.NewInitHandler(),
		mw.WithLogging[app.InitCmd]("app", logger),
	)

	checkHandler := mw.Chain(
		app.NewCheckHandler(),
		mw.WithLogging[app.CheckComamnd]("app", logger),
	)

	var err error
	switch subcmd {
	case "run":
		err = runHandler(ctx, app.RunCmd{})
	case "init":
		err = initHandler(ctx, app.InitCmd{})
	case "check":
		err = checkHandler(ctx, app.CheckComamnd{})
	default:
		usage()
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
