// asm is a standalone port of the buildx tool
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"

	// _ "github.com/docker/buildx/driver/kubernetes"

	"github.com/moby/buildkit/util/tracing/detect"
	_ "github.com/moby/buildkit/util/tracing/detect/delegated"
	_ "github.com/moby/buildkit/util/tracing/env"
	"go.opentelemetry.io/otel"

	"github.com/robertgzr/asm/config"
	"github.com/robertgzr/asm/version"
)

type ctxKeyConfig struct{}

func init() {
	detect.ServiceName = "asm"

	// do not log tracing errors to stdio
	otel.SetErrorHandler(skipErrors{})
}

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
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "",
			Usage:   "config file with worker infos",
		},
	}

	app.Commands = []*cli.Command{
		bakeCommand,
		genCommand,
		// serveCommand,
		// ctlCommand,
		nodesCommand,
	}

	app.Before = func(cx *cli.Context) error {
		if cx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("debug output enabled")
		}

		cfg, err := config.Load(cx.String("config"))
		if err != nil {
			return errors.Wrap(err, "loading config")
		}
		cx.Context = context.WithValue(cx.Context, ctxKeyConfig{}, cfg)

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

type skipErrors struct{}

func (skipErrors) Handle(err error) {}
