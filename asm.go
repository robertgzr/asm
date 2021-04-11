package asm

import (
	"context"
	"os"

	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/build"
	"github.com/docker/buildx/store"
	"github.com/docker/buildx/util/progress"
	"github.com/sirupsen/logrus"

	"github.com/robertgzr/asm/util"
)

func Assemble(ctx context.Context, ng *store.NodeGroup, conf *bake.Config, targets []string, progressMode, contextHash string) error {
	logrus.WithFields(logrus.Fields{
		"node_group":  ng,
		"bake_config": conf,
	}).Debug("starting assembly")

	m, err := configToTargets(conf, targets)
	if err != nil {
		return err
	}
	logrus.WithField("config", m).Debug("resolved config")

	opts, err := bake.TargetsToBuildOpt(m)
	if err != nil {
		return err
	}
	logrus.WithField("opts", opts).Debug("resolved opts")

	dis, err := util.DriversForNodeGroup(ctx, ng, contextHash)
	if err != nil {
		return err
	}
	logrus.WithField("drivers", dis).Debug("resolved drivers")

	ctx2, cancel := context.WithCancel(context.TODO())
	defer cancel()
	pw := progress.NewPrinter(ctx2, os.Stderr, progressMode)

	_, err = build.Build(ctx, dis, opts, nil, nil, pw)
	return err
}

func configToTargets(c *bake.Config, targets []string) (map[string]*bake.Target, error) {
	m := map[string]*bake.Target{}
	for _, n := range targets {
		for _, n := range c.ResolveGroup(n) {
			t, err := c.ResolveTarget(n, nil)
			if err != nil {
				return nil, err
			}
			if t != nil {
				m[n] = t
			}
		}
	}
	return m, nil
}
