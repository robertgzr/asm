package asm

import (
	"context"
	"os"

	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/build"
	"github.com/docker/buildx/util/progress"
	"github.com/sirupsen/logrus"

	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"

	// _ "github.com/docker/buildx/driver/kubernetes"
	// _ "github.com/robertgzr/asm/driver/containerd"

	"github.com/robertgzr/asm/config"
)

func Assemble(ctx context.Context, ng *config.NodeGroup, targets map[string]*bake.Target, progressMode, contextHash string) error {
	logrus.WithFields(logrus.Fields{
		"progress": progressMode,
		"context":  contextHash,
		"nodes":    len(ng.Nodes),
	}).Debug("starting assembly")

	logrus.Debugf("provided node group: %#+v", ng.Nodes)

	bo, err := bake.TargetsToBuildOpt(targets, nil)
	if err != nil {
		return err
	}
	logrus.Debugf("resolved opts: %#+v", bo)

	dis, err := DriversForNodeGroup(ctx, ng, contextHash)
	if err != nil {
		return err
	}
	logrus.Debugf("resolved drivers: %#+v", dis)

	ctx2, cancel := context.WithCancel(context.TODO())
	defer cancel()

	printer := progress.NewPrinter(ctx2, os.Stderr, progressMode)
	defer func() {
		if printer != nil {
			err1 := printer.Wait()
			if err == nil {
				err = err1
			}
		}
	}()

	_, err = build.Build(ctx, dis, bo, nil, nil, printer)
	return err
}
