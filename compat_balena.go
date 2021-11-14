//go:build balena
// +build balena

package asm

import (
	"errors"
	"fmt"
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
	"sigs.k8s.io/yaml"
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
	dfInline = strings.ReplaceAll(dfInline, "%%BALENA_SERVICE_NAME%%", t.Name)

	if v, ok := os.LookupEnv("ASM_BALENA_APP_NAME"); ok {
		dfInline = strings.ReplaceAll(dfInline, "%%BALENA_APP_NAME%%", v)
	} else {
		lg.Warn("BALENA_APP_NAME undefined")
	}
	if v, ok := os.LookupEnv("ASM_BALENA_RELEASE_HASH"); ok {
		dfInline = strings.ReplaceAll(dfInline, "%%BALENA_RELEASE_HASH%%", v)
	} else {
		lg.Warn("BALENA_RELEASE_HASH undefined")
	}

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

type BalenaYML struct {
	BuildSecrets struct {
		Global []struct {
			Source      string `yaml:"source"`
			Destination string `yaml:"dest"`
		} `yaml:"global"`
		Services map[string][]struct {
			Source      string `yaml:"source"`
			Destination string `yaml:"dest"`
		} `yaml:"services"`
	} `yaml:"build-secrets"`

	BuildVariables struct {
		Global   []string            `yaml:"global"`
		Services map[string][]string `yaml:"services"`
	} `yaml:"build-variables"`
}

func processBalenaYML(m map[string]*bake.Target, composeFilePath string) error {
	composeFileAbs, err := filepath.Abs(composeFilePath)
	if err != nil {
		return fmt.Errorf("detecting project root at %s: %w", composeFilePath, err)
	}
	balenaYMLPath := filepath.Join(filepath.Dir(composeFileAbs), ".balena/balena.yml")
	b, err := ioutil.ReadFile(balenaYMLPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logrus.
				WithField("balena", "balena.yml").
				Debug("no balena.yml found")
			return nil
		}
		return fmt.Errorf("processing balena.yml: %w", err)
	}

	balenaDir := filepath.Dir(balenaYMLPath)

	var balenaYML BalenaYML
	if err := yaml.Unmarshal(b, &balenaYML); err != nil {
		return err
	}
	// balena.yml > build-secrets
	mkSecret := func(src, dst string) string {
		return fmt.Sprintf("id=%s,src=%s", dst, filepath.Join(balenaDir, "secrets", src))
	}
	for _, s := range balenaYML.BuildSecrets.Global {
		for _, t := range m {
			t.Secrets = append(t.Secrets, mkSecret(s.Source, s.Destination))
		}
	}
	for name, secrets := range balenaYML.BuildSecrets.Services {
		t := m[name]
		for _, s := range secrets {
			t.Secrets = append(t.Secrets, mkSecret(s.Source, s.Destination))
		}
		m[name] = t
	}

	// balena.yml > build-variables
	mkArg := func(args map[string]string, arg string) {
		parts := strings.SplitN(arg, "=", 2)
		args[parts[0]] = parts[1]
	}
	for _, a := range balenaYML.BuildVariables.Global {
		for _, t := range m {
			if t.Args == nil {
				t.Args = make(map[string]string, len(balenaYML.BuildVariables.Global))
			}
			mkArg(t.Args, a)
		}
	}
	for name, args := range balenaYML.BuildVariables.Services {
		t := m[name]
		if t.Args == nil {
			t.Args = make(map[string]string, len(args))
		}
		for _, a := range args {
			mkArg(t.Args, a)
		}
		m[name] = t
	}

	return nil
}

func parseBalena(m map[string]*bake.Target, files []bake.File) error {
	if len(m) == 0 || len(files) == 0 {
		return nil
	}

	for name, t := range m {
		if t.Name == "" {
			t.Name = name
		}

		if t.Dockerfile == nil || *t.Dockerfile == "Dockerfile.template" {
			if err := processDockerfileTemplate(t); err != nil {
				return fmt.Errorf("processing dockerfile template: %w", err)
			}
		}
	}

	if err := processBalenaYML(m, files[0].Name); err != nil {
		return fmt.Errorf("processing balena.yml: %w", err)
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
