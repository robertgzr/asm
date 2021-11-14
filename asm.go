package asm

import (
	"context"

	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/build"
	"github.com/docker/buildx/util/progress"
	"github.com/robertgzr/asm/config"
	"github.com/sirupsen/logrus"

	_ "github.com/docker/buildx/driver/docker"
	_ "github.com/docker/buildx/driver/docker-container"
	// _ "github.com/docker/buildx/driver/kubernetes"
	// _ "github.com/robertgzr/asm/driver/containerd"
)

func Assemble(ctx context.Context, dis []build.DriverInfo, targets map[string]*bake.Target, inp *bake.Input, printer *progress.Printer) error {
	logrus.WithField("nodes", len(dis)).Debug("starting assembly")

	configDir, err := config.ConfigDir()
	if err != nil {
		return err
	}

	bo, err := bake.TargetsToBuildOpt(targets, inp)
	if err != nil {
		return err
	}
	if err := resolveBalena(bo, targets); err != nil {
		return err
	}
	logrus.Debugf("resolved opts: %+v", bo)

	_, err = build.Build(ctx, dis, bo, nil, configDir, printer)
	return err
}
