package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/urfave/cli/v3"
	"oci-runtime/internal/app"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/infrastructure/linux/mount"
	"oci-runtime/internal/infrastructure/linux/network"
	"oci-runtime/internal/infrastructure/linux/ns"
	"oci-runtime/internal/infrastructure/linux/proc"
	"oci-runtime/internal/infrastructure/linux/rs"
	"oci-runtime/internal/infrastructure/technical/config"
	"oci-runtime/internal/infrastructure/technical/logging"
	"os"
	"sort"
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `oci-runtime
Usage:
  oci-runtime run
  oci-runtime init
  oci-runtime check
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

	cmd := NewCmd(
		Actions{
			Run: func() mw.HandlerFunc[app.RunCmd] {
				return mw.Chain(
					app.NewRunHandler(),
					mw.WithLogging[app.RunCmd]("app", logger),
				)
			},
			Create: func() mw.HandlerFunc[app.CreateCmd] {
				return mw.Chain(
					app.NewCreateHandler(),
					mw.WithLogging[app.CreateCmd]("app", logger),
				)
			},
			Start: func() mw.HandlerFunc[app.StartCmd] {
				return mw.Chain(
					app.NewStartHandler(),
					mw.WithLogging[app.StartCmd]("app", logger),
				)
			},
			Init: func() mw.HandlerFunc[app.InitCmd] {
				return mw.Chain(
					app.NewInitHandler(app.Ports{
						Mount: mount.NewManager(),
						NS:    ns.NewManager(),
						Root:  rs.NewManager(),
						Proc:  proc.NewManager(),
						Net:   network.NewManager(),
					}),
					mw.WithLogging[app.InitCmd]("app", logger),
				)
			},
			Check: func() mw.HandlerFunc[app.CheckComamnd] {
				return mw.Chain(
					app.NewCheckHandler(),
					mw.WithLogging[app.CheckComamnd]("app", logger),
				)
			},
		},
	)

	sort.Sort(cli.FlagsByName(cmd.Flags))
	err := cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
