package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/util/progress"
	"github.com/docker/buildx/util/tracing"
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
		&cli.StringSliceFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "build definition file",
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
		&cli.StringSliceFlag{
			Name:  "nodes",
			Usage: "overwrite the build nodes",
		},
	},
	Action: func(cx *cli.Context) (err error) {
		cfg := cx.Context.Value(ctxKeyConfig{}).(config.NodeGroup)
		if len(cfg.Nodes) == 0 {
			return fmt.Errorf("missing node configuration")
		}

		logrus.Debugf("node configuration: %+v", cfg)

		if nodes := cx.StringSlice("nodes"); len(nodes) != 0 {
			var filtered config.NodeGroup
			for _, n := range cfg.Nodes {
				for _, name := range nodes {
					if n.Name == name {
						filtered.Nodes = append(filtered.Nodes, n)
					}
				}
			}
			if len(filtered.Nodes) == 0 {
				return fmt.Errorf("no nodes left")
			}
			cfg = filtered
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctx, end, err := tracing.TraceCurrentCommand(ctx, "bake")
		if err != nil {
			return err
		}
		defer func() {
			end(err)
		}()

		ctx2, cancelPrinter := context.WithCancel(context.TODO())
		defer cancelPrinter()
		printer := progress.NewPrinter(ctx2, os.Stderr, cx.String("progress"))
		defer func() {
			if printer != nil {
				err1 := printer.Wait()
				if err == nil {
					err = err1
				}
			}
		}()

		targets := []string{"default"}
		if cx.Args().Present() {
			targets = cx.Args().Slice()
		}

		contextPathHash, _ := os.Getwd()
		dis, err := asm.DriversForNodeGroup(ctx, &cfg, contextPathHash)
		if err != nil {
			return err
		}
		logrus.Debugf("resolved drivers: %+v", dis)

		var (
			url      string
			defaults = map[string]string{
				"BAKE_LOCAL_PLATFORM": platforms.DefaultString(),
				"BAKE_CMD_CONTEXT":    "cwd://",
			}
		)

		if len(targets) > 0 && bake.IsRemoteURL(targets[0]) {
			url = targets[0]
			targets = targets[1:]
			if len(targets) > 0 && bake.IsRemoteURL(targets[0]) {
				defaults["BAKE_CMD_CONTEXT"] = targets[0]
				targets = targets[1:]
			}
		}

		// set a default target
		if len(targets) == 0 {
			targets = []string{"default"}
		}

		var (
			files []bake.File
			inp   *bake.Input
		)

		if url != "" {
			logrus.WithField("url", url).Debugf("pulling remote files")
			files, inp, err = bake.ReadRemoteFiles(ctx, dis, url, cx.StringSlice("file"), printer)
		} else {
			files, err = bake.ReadLocalFiles(cx.StringSlice("file"))
		}
		if err != nil {
			return fmt.Errorf("reading files failed: %w", err)
		}

		logrus.Debugf("resolved files: %+v", files)

		m, err := asm.ReadTargets(ctx, files, targets, cx.StringSlice("set"), defaults)
		if err != nil {
			return err
		}

		logrus.Debugf("resolved targets: %+v", m)

		if cx.Bool("print") {
			// end printer early
			if err := printer.Wait(); err != nil {
				return fmt.Errorf("printer failed: %w", err)
			}
			printer = nil

			dt, err := json.MarshalIndent(map[string]map[string]*bake.Target{"target": m}, "", "   ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cx.App.Writer, string(dt))
			return nil
		}

		if err := asm.Assemble(ctx, dis, m, inp, printer); err != nil {
			return fmt.Errorf("assembly failed: %w", err)
		}
		return nil
	},
}
