// +build balena_compat

package asm

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/opencontainers/image-spec/specs-go/v1"
	// "github.com/docker/buildx/bake"
	"github.com/sirupsen/logrus"
)

var balenalibRE = regexp.MustCompile(`FROM.*balenalib\/.*`)

func balenaCompat(m targetMap) error {
	var (
		platform = platforms.DefaultSpec()
		arch     = ""
		machine  = ""
	)
	if v, ok := os.LookupEnv("ASM_BALENA_ARCH"); ok {
		arch = v
	}
	if v, ok := os.LookupEnv("ASM_BALENA_MACHINE_NAME"); ok {
		machine = v
	}

	for _, t := range m {
		lg := logrus.WithField("context", *t.Context)
		lg.Debugf("trying Dockerfile.template")
		b, err := ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile.template"))
		if os.IsNotExist(err) {
			lg.Debugf("trying Dockerfile.%s", arch)
			b, err = ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile."+arch))
			if os.IsNotExist(err) {
				lg.Debugf("trying Dockerfile.%s", machine)
				b, err = ioutil.ReadFile(filepath.Join(*t.Context, "Dockerfile."+machine))
				if os.IsNotExist(err) {
					continue
				}
			}
		}

		if len(t.Platforms) == 1 {
			platform = platforms.Normalize(platforms.MustParse(t.Platforms[0]))
		} else if len(t.Platforms) > 1 {
			return errors.New("multiple platforms not supported")
		}

		if arch == "" {
			if v, ok := toArch(platform); ok {
				arch = v
			} else {
				return errors.New("unable to translate into BALENA_ARCH")
			}
		}
		if machine == "" {
			if v, ok := toMachine(platform); ok {
				machine = v
			} else {
				return errors.New("unable to translate into BALENA_MACHINE_NAME")
			}
		}

		logrus.
			WithField("BALENA_ARCH_NAME", arch).
			WithField("BALENA_MACHINE_NAME", machine).
			Infof("running balena compat for target %s", t.Name)

		dfInline := string(b)
		dfInline = strings.ReplaceAll(dfInline, "%%BALENA_MACHINE_NAME%%", machine)
		dfInline = strings.ReplaceAll(dfInline, "%%BALENA_ARCH%%", arch)
		t.DockerfileInline = &dfInline
		empty := ""
		t.Dockerfile = &empty

		// TODO remove this
		if balenalibRE.Match([]byte(dfInline)) {
			logrus.Warn("NOTICE: balenalib images are broken when used with the platform option.")
			logrus.Warn("        Applying fix to unblock build.")
			t.Platforms = []string{}
		}
	}

	return nil
}

func toMachine(spec v1.Platform) (string, bool) {
	switch {
	case spec.Architecture == "amd64":
		return "intel-nuc", true
	case spec.Architecture == "i386":
		return "intel-edison", true
	case spec.Architecture == "arm64":
		return "raspberrypi4-64", true
	case spec.Architecture == "arm" && spec.Variant == "v7":
		return "raspberrypi3", true
	case spec.Architecture == "arm" && spec.Variant == "v6":
		return "raspberry-pi", true
	default:
		return "", false
	}
}

func toArch(spec v1.Platform) (string, bool) {
	switch spec.Architecture {
	case "amd64":
		return "amd64", true
	case "i386":
		return "i386", true
	case "arm64":
		return "aarch64", true
	case "arm":
		if spec.Variant == "v7" {
			return "armv7hf", true
		}
		if spec.Variant == "v6" {
			return "rpi", true
		}
		return "", false
	default:
		return "", false
	}
}
