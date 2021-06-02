// asm is a standalone port of the buildx tool
package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"

	// _ "github.com/docker/buildx/driver/kubernetes"

	"github.com/robertgzr/asm/version"
)

func main() {
	cli.VersionPrinter = func(cx *cli.Context) {
		fmt.Println(cx.App.Name, cx.App.Version, version.Revision)
	}
	app := cli.NewApp()
	app.Name = "asm"
	app.Usage = "the standalone buildx"
	app.Version = version.Version

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "be more verbose",
		},
	}

	app.Commands = []*cli.Command{
		bakeCommand,
		genCommand,
		// serveCommand,
		// ctlCommand,
	}

	app.Before = func(cx *cli.Context) error {
		if cx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("debug output enabled")
		}
		return nil
	}

	app.Action = func(cx *cli.Context) error {
		return cli.ShowAppHelp(cx)
	}

	if err := app.Run(os.Args); err != nil {
		logrus.SetOutput(os.Stderr)
		logrus.WithError(err).Fatalf("%s stopped", app.Name)
	}
}
