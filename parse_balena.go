// +build balena_compat

package asm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/buildx/bake"
	"github.com/sirupsen/logrus"
)

func balenaCompat(fn string, t *bake.Target) {
	context := filepath.Join(filepath.Dir(fn), *t.Context)
	if _, err := os.Stat(filepath.Join(context, *t.Dockerfile)); err == nil {
		return
	}

	var (
		arch    = platforms.DefaultSpec().Architecture
		machine = arch
	)
	if v, ok := os.LookupEnv("ASM_BALENA_ARCH"); ok {
		arch = v
	}
	if v, ok := os.LookupEnv("ASM_BALENA_MACHINE_NAME"); ok {
		machine = v
	}

	switch {
	case arch == "rpi" || machine == "raspberrypi":
		t.Platforms = []string{"linux/arm/v6"}
		arch = "rpi"
	case arch == "armv7hf" || isArm7(machine):
		t.Platforms = []string{"linux/arm/v7"}
		arch = "armv7hf"
	case arch == "aarch64" || isArm64(machine):
		t.Platforms = []string{"linux/arm64"}
		arch = "aarch64"
	default:
		t.Platforms = []string{"linux/" + arch}
	}

	logrus.
		WithField("arch", arch).
		WithField("machine", machine).
		Warnf("parse: running balena compat for service %q", t.Name)

	lg := logrus.WithField("context", context)
	lg.Debugf("parse: trying Dockerfile.template")
	b, err := ioutil.ReadFile(filepath.Join(context, "Dockerfile.template"))
	if os.IsNotExist(err) {
		lg.Debugf("parse: trying Dockerfile.%s", arch)
		b, err = ioutil.ReadFile(filepath.Join(context, "Dockerfile."+arch))
		if os.IsNotExist(err) {
			lg.Debugf("parse: trying Dockerfile.%s", machine)
			b, err = ioutil.ReadFile(filepath.Join(context, "Dockerfile."+machine))
			if os.IsNotExist(err) {
				panic("missing dockerfile")
			}
		}
	}

	dfInline := string(b)
	dfInline = strings.ReplaceAll(dfInline, "%%BALENA_MACHINE_NAME%%", machine)
	dfInline = strings.ReplaceAll(dfInline, "%%BALENA_ARCH_NAME%%", arch)
	t.DockerfileInline = &dfInline
	empty := ""
	t.Dockerfile = &empty
	t.Context = &context
}

var arm7machines = []string{
	"raspberrypi2",
	"raspberrypi3",
}

func isArm7(machine string) bool {
	for _, m := range arm7machines {
		if machine == m {
			return true
		}
	}
	return false
}

var arm64machines = []string{
	"raspberrypi3-64",
	"raspberrypi4",
}

func isArm64(machine string) bool {
	for _, m := range arm64machines {
		if machine == m {
			return true
		}
	}
	return false
}
