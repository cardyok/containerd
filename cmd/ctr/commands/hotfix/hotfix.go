package hotfix

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/containerd/containerd/cmd/ctr/commands"
)

// Command is a cli command to apply hotfix on containerd main process
var Command = cli.Command{
	Name:  "hotfix",
	Usage: "miscellaneous hotfix commands",
	Subcommands: cli.Commands{
		ChangeLogLevelCommand,
	},
}

var ChangeLogLevelCommand = cli.Command{
	Name:    "change-log-level",
	Aliases: []string{"clog"},
	Usage:   "change log level of containerd main process",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "level, l",
			Usage: "log level to use, e.g. info",
		},
	},
	Action: func(context *cli.Context) error {
		client, ctx, cancel, err := commands.NewClient(context)
		if err != nil {
			return err
		}
		defer cancel()
		logLevel := context.String("level")
		if logLevel == "" {
			fmt.Println("no level specified, exiting..")
			return nil
		}
		err = client.ChangeLogLevel(ctx, logLevel)
		return err
	},
}
