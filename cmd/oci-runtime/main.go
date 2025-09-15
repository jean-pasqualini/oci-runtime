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
	"oci-runtime/internal/infrastructure/transport/ipc"
	"os"
	"sort"
)

func main() {
	var root string
	var logpath string
	var logformat string
	flag.StringVar(&root, "root", "", "root")
	flag.StringVar(&logpath, "log", "/tmp/oci-runtime-log.json", "log")
	flag.StringVar(&logformat, "log-format", "json", "logformat")
	flag.Parse()

	_ = config.Load()
	logger, err := logging.New(flag.Arg(0), logpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx := context.Background()

	cmd := NewCmd(
		Actions{
			Create: func() mw.HandlerFunc[app.CreateCmd] {
				return mw.Chain(
					app.NewCreateHandler(ipc.NewSyncPipe),
					mw.WithLogging[app.CreateCmd]("app", logger),
				)
			},
			Start: func() mw.HandlerFunc[app.StartCmd] {
				return mw.Chain(
					app.NewStartHandler(ipc.NewSyncPipe),
					mw.WithLogging[app.StartCmd]("app", logger),
				)
			},
			Init: func() mw.HandlerFunc[app.InitCmd] {
				return mw.Chain(
					app.NewInitHandler(app.Ports{
						Mount:      mount.NewManager(),
						NS:         ns.NewManager(),
						Root:       rs.NewManager(),
						Proc:       proc.NewManager(),
						Net:        network.NewManager(),
						IpcFactory: ipc.NewSyncPipe,
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
	err = cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
