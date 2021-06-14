package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/buildx/bake"
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
		&cli.StringSliceFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "build definition file",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "set",
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
		cfg := cx.Context.Value(ctxKeyConfig{}).(config.NodeGroup)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		targets := []string{"default"}
		if cx.Args().Present() {
			targets = cx.Args().Slice()
		}

		files, err := bake.ReadLocalFiles(cx.StringSlice("file"))
		if err != nil {
			return err
		}

		m, err := asm.ReadTargets(ctx, files, targets, cx.StringSlice("set"))
		if err != nil {
			return err
		}

		if cx.Bool("print") {
			dt, err := json.MarshalIndent(map[string]map[string]*bake.Target{"target": m}, "", "   ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cx.App.Writer, string(dt))
			return nil
		}

		contextPathHash, _ := os.Getwd()
		return asm.Assemble(ctx, &cfg, m, cx.String("progress"), contextPathHash)
	},
}
