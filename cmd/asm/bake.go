package main

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/docker/buildx/bake"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/robertgzr/asm"
	"github.com/robertgzr/asm/config"
)

var bakeCommand = &cli.Command{
	Name:        "bake",
	Aliases:     []string{"f"},
	Usage:       "bake [TARGET...]",
	Description: "build from a file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "config file with worker infos",
			Value:   "config.json",
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "build definition file",
			Value:   "compose.yaml",
		},
		&cli.StringFlag{
			Name:  "set",
			Value: "",
			Usage: "override target value (eg: targetpattern.key=value)",
		},
		&cli.BoolFlag{
			Name:  "print",
			Usage: "print options without building",
		},
		&cli.StringFlag{
			Name:  "progress",
			Value: "auto",
			Usage: "set type of progress output (auto, plain, tty)",
		},
	},
	Action: func(cx *cli.Context) error {
		cfg, err := config.Load(cx.String("config"))
		if err != nil {
			return err
		}

		if cx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("debug output enabled")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		fn := cx.String("file")

		targets := []string{"default"}
		if cx.Args().Present() {
			targets = cx.Args().Slice()
		}

		c, err := asm.ParseConfig(fn)
		if err != nil {
			return err
		}

		if cx.Bool("print") {
			return printConfig(c, cx.App.Writer)
		}

		contextPathHash, _ := os.Getwd()
		return asm.Assemble(ctx, &cfg, c, targets, cx.String("progress"), contextPathHash)
	},
}

func printConfig(c *bake.Config, w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}
