package main

import (
	"encoding/json"
	"os"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/buildx/store"
	cli "github.com/urfave/cli/v2"
)

var genCommand = &cli.Command{
	Name:    "generate",
	Aliases: []string{"gen"},
	Usage:   "generate files",
	// Description: "",
	Action: func(cx *cli.Context) error {
		return cli.ShowCommandHelp(cx, cx.Command.Name)
	},
	Subcommands: []*cli.Command{
		genDockerCommand,
	},
}

var genDockerCommand = &cli.Command{
	Name:  "docker",
	Usage: "generate docker node group configs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "host",
			Aliases:  []string{"H"},
			Usage:    "docker daemon address",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "platform",
			Usage: "platforms supported by this docker daemon",
			Value: cli.NewStringSlice(platforms.DefaultString()),
		},
	},
	Action: func(cx *cli.Context) error {
		node := store.Node{
			Name:     "docker daemon",
			Endpoint: cx.String("host"),
		}
		for _, pl := range cx.StringSlice("platform") {
			spec, err := platforms.Parse(pl)
			if err != nil {
				return err
			}
			node.Platforms = append(node.Platforms, spec)
		}
		config := store.NodeGroup{
			Name:   "default",
			Driver: "docker",
			Nodes:  []store.Node{node},
		}
		err := json.NewEncoder(os.Stdout).Encode(config)
		return err
	},
}
