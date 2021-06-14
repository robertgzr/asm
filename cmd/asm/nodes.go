package main

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/containerd/containerd/platforms"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"

	"github.com/robertgzr/asm/config"
)

const keyConfig = "io.robertgzr.asm.config"

var nodesCommand = &cli.Command{
	Name:    "nodes",
	Aliases: []string{},
	Usage:   "interact with build nodes",
	Before: func(cx *cli.Context) error {
		cfg, err := config.Load(cx.String("config"))
		if err != nil {
			return errors.Wrap(err, "loading config")
		}
		cx.Context = context.WithValue(cx.Context, keyConfig, cfg)
		return nil
	},
	Action: func(cx *cli.Context) error {
		return listNodes(cx)
	},
	Subcommands: []*cli.Command{listNodesCommand},
}

var listNodesCommand = &cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "list build nodes",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "platform",
			Usage: "platforms supported by this docker daemon",
			Value: cli.NewStringSlice(platforms.DefaultString()),
		},
	},
	Action: listNodes,
}

func formatPlatformArray(array []v1.Platform) string {
    var ps []string
    for _, p := range array {
      ps = append(ps, platforms.Format(p))
    }
    return strings.Join(ps, ",")
}

func listNodes(cx *cli.Context) error {
	cfg := cx.Context.Value(keyConfig).(config.NodeGroup)

	tw := tabwriter.NewWriter(cx.App.Writer, 0, 4, 4, ' ', tabwriter.TabIndent)

	fmt.Fprintf(tw, "NAME\tDRIVER\tENDPOINT\tPLATFORMS\n")
	for _, n := range cfg.Nodes {
		defer tw.Flush()
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", n.Name, n.Driver, n.Endpoint, formatPlatformArray(n.Platforms))
	}

	return nil
}
