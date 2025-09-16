package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/app"
	"oci-runtime/internal/app/mw"
	"syscall"
)

type Actions struct {
	Create func() mw.HandlerFunc[app.CreateCmd]
	Start  func() mw.HandlerFunc[app.StartCmd]
	Init   func() mw.HandlerFunc[app.InitCmd]
	Check  func() mw.HandlerFunc[app.CheckComamnd]
}

const usage = `Open Container Initiative runtime

oci-runtime is a command line client for running applications packaged according to
the Open Container Initiative (OCI) format and is a compliant implementation of the
Open Container Initiative specification.

Containers are configured using bundles. A bundle for a container is a directory
that includes a specification file named "config.json" and a root filesystem.
The root filesystem contains the contents of the container.

To start a new instance of a container:

    # oci-runtime --root <path> run --bundle <path> <container-id>
`

func requireExactArgs(n int, hint string) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.NArg() != n {
			return ctx, cli.Exit(fmt.Sprintf("expected %d args: %s", n, hint), 2)
		}
		return ctx, nil
	}
}

func NewCmd(actions Actions) cli.Command {
	return cli.Command{
		Name:      "oci-runtime",
		UsageText: usage,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "root", Usage: "internal state folder path", Required: true},
			&cli.StringFlag{Name: "log", Usage: "where the runtime logs are stored", Required: false},
			&cli.StringFlag{Name: "log-format", Usage: "what is the log format", Required: false},
		},
		Commands: []*cli.Command{
			{
				Name:      "run",
				Usage:     "run a contaienr",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bundle", Usage: "bundle folder path: contains the config.json + rootfs", Required: true},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
						return err
					}

					if err := actions.Create()(ctx, app.CreateCmd{
						Name:          cmd.StringArg("name"),
						MetadataRoot:  cmd.String("root"),
						BundleRoot:    cmd.String("bundle"),
						LogPath:       cmd.String("log"),
						LogFormat:     cmd.String("log-format"),
						PidFile:       cmd.String("pid-file"),
						ConsoleSocket: cmd.String("console-socket"),
					}); err != nil {
						return err
					}

					if err := actions.Start()(ctx, app.StartCmd{
						Name:         cmd.StringArg("name"),
						MetadataRoot: cmd.String("root"),
					}); err != nil {
						return err
					}

					var waitInit syscall.WaitStatus
					_, err := syscall.Wait4(-1, &waitInit, 0, nil)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "create",
				Usage: "create a contaienr",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "bundle", Usage: "bundle folder path: contains the config.json + rootfs", Required: true},
					&cli.StringFlag{Name: "pid-file", Usage: "pid file path: contains the init pid", Required: false},
					&cli.StringFlag{Name: "console-socket", Usage: "console socket file: contains the pty", Required: false},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					err := actions.Create()(ctx, app.CreateCmd{
						Name:          cmd.StringArg("name"),
						MetadataRoot:  cmd.String("root"),
						BundleRoot:    cmd.String("bundle"),
						LogPath:       cmd.String("log"),
						LogFormat:     cmd.String("log-format"),
						PidFile:       cmd.String("pid-file"),
						ConsoleSocket: cmd.String("console-socket"),
					})
					if err != nil {
						// Debug purpose
						/**
						data, debugErr := os.ReadFile(cmd.String("log")) // reads the entire file
						if debugErr != nil {
							return err
						}
						fmt.Println(string(data))
						*/
						return err
					}
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "start a contaienr",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name"},
				},
				Before: requireExactArgs(1, "<name>"),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return actions.Start()(ctx, app.StartCmd{
						Name:         cmd.StringArg("name"),
						MetadataRoot: cmd.String("root"),
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
