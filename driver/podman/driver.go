package podman

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/containers/toolbox/pkg/podman"
	"github.com/containers/toolbox/pkg/shell"
	"github.com/docker/buildx/driver"
	"github.com/docker/buildx/driver/bkimage"
	"github.com/docker/buildx/util/progress"
	"github.com/moby/buildkit/client"
	"github.com/sirupsen/logrus"

	asmdriver "github.com/robertgzr/asm/driver"
)

var volumeStateSuffix = "_state"

type Driver struct {
	factory driver.Factory
	driver.InitConfig
	image string
}

func (d *Driver) Factory() driver.Factory {
	return d.factory
}

func (d *Driver) Bootstrap(ctx context.Context, l progress.Logger) error {
	return progress.Wrap("[internal] booting buildkit", l, func(sub progress.SubLogger) error {
		info, err := d.Info(ctx)
		if err != nil {
			return err
		}

		if info.Status == driver.Running {
			return nil
		}

		if info.Status == driver.Inactive {
			if err := d.create(ctx, sub); err != nil {
				return err
			}
		}

		return sub.Wrap("starting container", func() error {
			if err := d.start(ctx, sub); err != nil {
				return err
			}
			return nil
		})
	})
}

func (d *Driver) create(ctx context.Context, l progress.SubLogger) error {
	imageName := "docker.io/" + bkimage.DefaultImage
	if d.image != "" {
		imageName = d.image
	}

	if err := l.Wrap("pulling image "+imageName, func() error {
		return podman.Pull(imageName)
	}); err != nil {
		l.Wrap("pulling failed, using local image "+imageName, func() error { return nil })
	}

	createArgs := []string{
		"--log-level", podman.LogLevel.String(),
		"create",
		"--name", d.Name,
		"--privileged",
		"--userns=host",
		"--mount=type=volume,source=" + d.Name + volumeStateSuffix + ",target=/var/lib/buildkit",
	}
	// TODO env
	// TODO network
	createArgs = append(createArgs, imageName)

	return l.Wrap("creating container "+d.Name, func() error {
		var stderr strings.Builder
		if err := shell.Run("podman", nil, nil, &stderr, createArgs...); err != nil {
			return fmt.Errorf("failed to create container: %s", stderr.String())
		}
		return nil
	})
}

func (d *Driver) start(ctx context.Context, l progress.SubLogger) error {
	var stderr strings.Builder
	if err := podman.Start(d.Name, &stderr); err != nil {
		return errors.New(stderr.String())
	}
	return nil
}

func (d *Driver) Info(ctx context.Context) (*driver.Info, error) {
	containers, err := podman.GetContainers("--all", "--filter=name="+d.Name)
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		logrus.Debug("No containers not found, marking driver inactive")
		return &driver.Info{
			Status: driver.Inactive,
		}, nil
	}
	if containers[0]["Status"] != "running" {
		logrus.Debug("Container found but not running, marking driver stopped")
		return &driver.Info{
			Status: driver.Stopped,
		}, nil
	}
	return &driver.Info{
		Status: driver.Running,
	}, nil
}

func (d *Driver) Stop(ctx context.Context, force bool) error {
	stopArgs := []string{
		"--log-level", podman.LogLevel.String(),
		"stop",
		"--ignore",
		d.Name,
	}

	var stderr strings.Builder
	if err := shell.Run("podman", nil, nil, &stderr, stopArgs...); err != nil {
		return fmt.Errorf("failed to stop container: %s", stderr.String())
	}
	return nil
}

func (d *Driver) Rm(ctx context.Context, force bool, rmVolume bool) error {
	info, err := d.Info(ctx)
	if err != nil {
		return err
	}
	if info.Status == driver.Inactive {
		return nil
	}
	containers, err := podman.GetContainers(d.Name)
	if err != nil {
		return err
	}
	if err := podman.RemoveContainer(d.Name, force); err != nil {
		return err
	}
	logrus.Warnf("%+v", containers[0]["Mounts"])
	// for _, v := range containers[0]["Mounts"] {
	// 	logrus.Warnf("%+v", v)
	// 	m, ok := v.(map[string]interface{})
	// 	if !ok {
	// 		panic("not a volume")
	// 	}
	// 	if m["Type"] == "volume" && m["Source"] == d.Name+volumeStateSuffix && rmVolume {
	// 		return errors.New("volume removal not implemented")
	// 	}
	// }
	return nil
}

func (d *Driver) exec(ctx context.Context, command []string) (*os.Process, net.Conn, error) {
	execArgs := []string{
		"--log-level", podman.LogLevel.String(),
		"exec",
		"--interactive",
		d.Name,
	}
	execArgs = append(execArgs, command...)

	cmd := exec.Command("podman", execArgs...)
	inp, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	outp, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	conn, err := asmdriver.NewStdioConn(ctx, inp, outp)
	return cmd.Process, conn, err
}

func (d *Driver) Client(ctx context.Context) (*client.Client, error) {
	p, conn, err := d.exec(ctx, []string{"buildctl", "dial-stdio"})
	if err != nil {
		return nil, err
	}
	go func(p *os.Process) {
		logrus.Warnf("%+v", p)
		ps, err := p.Wait()
		if err != nil {
			panic(err)
		}
		logrus.Warnf("%+v", ps)
	}(p)
	return client.New(ctx, "", client.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
		return conn, nil
	}))
}

func (d *Driver) Features() map[driver.Feature]bool {
	return map[driver.Feature]bool{
		driver.OCIExporter:    false,
		driver.DockerExporter: false,
		driver.CacheExport:    false,
		driver.MultiPlatform:  false,
	}
}

func (d *Driver) IsMobyDriver() bool {
	return false
}

func (d *Driver) Config() driver.InitConfig {
	return d.InitConfig
}
