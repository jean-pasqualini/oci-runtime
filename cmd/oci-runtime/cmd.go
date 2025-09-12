package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	"oci-runtime/internal/app"
	"oci-runtime/internal/app/mw"
)

type Actions struct {
	Run    func() mw.HandlerFunc[app.RunCmd]
	Create func() mw.HandlerFunc[app.CreateCmd]
	Start  func() mw.HandlerFunc[app.StartCmd]
	Init   func() mw.HandlerFunc[app.InitCmd]
	Check  func() mw.HandlerFunc[app.CheckComamnd]
}

func requireExactArgs(n int, hint string) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.NArg() != n {
			return ctx, cli.Exit(fmt.Sprintf("expected %d args: %s", n, hint), 2)
		}
		return ctx, nil
	}
}

func NewCmd(actions Actions) *cli.Command {
	return &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run a contaienr",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "root", Usage: "internal state folder path", Required: true},
					&cli.StringFlag{Name: "bundle", Usage: "bundle folder path: contains the config.json + rootfs", Required: true},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return actions.Run()(ctx, app.RunCmd{
						Name:          cmd.StringArg("name"),
						StatePath:     cmd.String("root"),
						BundlePath:    cmd.String("bundle"),
						LogPath:       cmd.String("log"),
						LogFormat:     cmd.String("log-format"),
						PidFile:       cmd.String("pid-file"),
						ConsoleSocket: cmd.String("console-socket"),
					})
				},
			},
			{
				Name:  "create",
				Usage: "create a contaienr",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "root", Usage: "internal state folder path", Required: true},
					&cli.StringFlag{Name: "bundle", Usage: "bundle folder path: contains the config.json + rootfs", Required: true},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return actions.Create()(ctx, app.CreateCmd{
						Name:          cmd.StringArg("name"),
						StatePath:     cmd.String("root"),
						BundlePath:    cmd.String("bundle"),
						LogPath:       cmd.String("log"),
						LogFormat:     cmd.String("log-format"),
						PidFile:       cmd.String("pid-file"),
						ConsoleSocket: cmd.String("console-socket"),
					})
				},
			},
			{
				Name:  "start",
				Usage: "start a contaienr",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "root", Usage: "internal state folder path", Required: true},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return actions.Start()(ctx, app.StartCmd{
						Name:      cmd.StringArg("name"),
						StatePath: cmd.String("root"),
					})
				},
			},
			{
				Name:   "init",
				Hidden: true,
				Usage:  "init (internal)",
				Action: func(ctx context.Context, command *cli.Command) error {
					return actions.Init()(ctx, app.InitCmd{})
				},
			},
			{
				Name:  "check",
				Usage: "is the machine able to run it",
				Action: func(ctx context.Context, command *cli.Command) error {
					return actions.Check()(ctx, app.CheckComamnd{})
				},
			},
		},
	}
}
