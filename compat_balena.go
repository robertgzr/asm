//go:build balena
// +build balena

package asm

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/buildx/bake"
	"github.com/docker/buildx/build"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

var balenalibRE = regexp.MustCompile(`FROM.*balenalib\/.*`)

func balenaPlatform(t *bake.Target) (machine, arch string, platform v1.Platform, err error) {
	var fromEnv bool
	machine, fromEnv = os.LookupEnv("ASM_BALENA_MACHINE_NAME")
	arch, fromEnv = os.LookupEnv("ASM_BALENA_ARCH")

	// detect platform
	if t.Platforms != nil && len(t.Platforms) > 0 {
		if len(t.Platforms) > 1 {
			err = errors.New("multiple platforms not supported")
			return
		}
		platform = platforms.Normalize(platforms.MustParse(t.Platforms[0]))
	} else {
		// fallback to build host platform
		platform = platforms.DefaultSpec()
	}

	if fromEnv {
		return machine, arch, platform, nil
	}

	machine, err = toMachine(platform)
	if err != nil {
		return
	}
	arch, err = toArch(platform)
	if err != nil {
		return
	}

	return machine, arch, platform, err
}

func toMachine(spec v1.Platform) (string, error) {
	switch {
	case spec.Architecture == "amd64":
		return "intel-nuc", nil
	case spec.Architecture == "i386":
		return "intel-edison", nil
	case spec.Architecture == "arm64":
		return "raspberrypi4-64", nil
	case spec.Architecture == "arm" && spec.Variant == "v7":
		return "raspberrypi3", nil
	case spec.Architecture == "arm" && spec.Variant == "v6":
		return "raspberry-pi", nil
	default:
		return "", errors.New("unable to translate into BALENA_MACHINE_NAME")
	}
}

func toArch(spec v1.Platform) (string, error) {
	switch spec.Architecture {
	case "amd64":
		return "amd64", nil
	case "i386":
		return "i386", nil
	case "arm64":
		return "aarch64", nil
	case "arm":
		if spec.Variant == "v7" {
			return "armv7hf", nil
		}
		if spec.Variant == "v6" {
			return "rpi", nil
		}
		fallthrough // to error below
	default:
		return "", errors.New("unable to translate into BALENA_ARCH")
	}
}

func processDockerfileTemplate(t *bake.Target) error {
	lg := logrus.
		WithField("balena", "template").
		WithField("context", *t.Context).
		WithField("target", t.Name)

	machine, arch, platform, err := balenaPlatform(t)
	if err != nil {
		return err
	}

	lg.Debugf("trying Dockerfile.template")
	b, err := ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile.template"))
	if errors.Is(err, os.ErrNotExist) {
		lg.Debugf("trying Dockerfile.%s", arch)
		b, err = ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile."+arch))
		if errors.Is(err, os.ErrNotExist) {
			lg.Debugf("trying Dockerfile.%s", machine)
			b, err = ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile."+machine))
			if errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}

	dfInline := string(b)
	dfInline = strings.ReplaceAll(dfInline, "%%BALENA_MACHINE_NAME%%", machine)
	dfInline = strings.ReplaceAll(dfInline, "%%BALENA_ARCH%%", arch)
	t.DockerfileInline = &dfInline
	empty := ""
	t.Dockerfile = &empty

	lg.
		WithField("PLATFORM", platforms.Format(platform)).
		WithField("BALENA_ARCH_NAME", arch).
		WithField("BALENA_MACHINE_NAME", machine).
		Info("processed template")

	// TODO remove this
	if balenalibRE.Match([]byte(dfInline)) {
		lg.Warn("NOTICE: balenalib images are broken when used with the platform option.")
		lg.Warn("        Applying fix to unblock build.")
		t.Platforms = []string{}
	}

	return nil
}

func parseBalena(m map[string]*bake.Target) error {
	for name, t := range m {
		if t.Name == "" {
			t.Name = name
		}

		if t.Dockerfile == nil || *t.Dockerfile == "Dockerfile.template" {
			if err := processDockerfileTemplate(t); err != nil {
				return err
			}
		}

		// TODO secrets
	}
	return nil
}

func resolveBalena(bo map[string]build.Options, m map[string]*bake.Target) error {
	lg := logrus.WithField("balena", "resolve")
	for target, o := range bo {
		machine, arch, platform, err := balenaPlatform(m[target])
		if err != nil {
			return err
		}

		o.Platforms = append(o.Platforms, platform)

		if o.BuildArgs == nil {
			o.BuildArgs = make(map[string]string, 3)
		}
		o.BuildArgs["BALENA_MACHINE_NAME"] = machine
		o.BuildArgs["BALENA_ARCH"] = arch
		o.BuildArgs["BALENA_SERVICE_NAME"] = target

		lg.
			WithField("target", target).
			WithField("PLATFORM", platforms.Format(platform)).
			WithField("BALENA_ARCH_NAME", arch).
			WithField("BALENA_MACHINE_NAME", machine).
			WithField("BALENA_SERVICE_NAME", target).
			Info("resolved build args")

		bo[target] = o
	}
	return nil
}
