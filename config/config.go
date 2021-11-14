package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/buildx/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	asmDir := filepath.Join(configDir, "asm")
	if _, err := os.Stat(asmDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(asmDir, os.ModeDir|os.ModePerm); err != nil {
			return "", err
		}
	}
	return asmDir, nil
}

type NodeGroup struct {
	Nodes []Node
}

type Node struct {
	store.Node
	Driver string
}

func load(dir string) (fp string) {
	fp = filepath.Join(dir, "asm.yml")
	if _, err := os.Stat(fp); err == nil {
		return fp
	}
	fp = filepath.Join(dir, "asm.yaml")
	if _, err := os.Stat(fp); err == nil {
		return fp
	}
	fp = filepath.Join(dir, "asm.json")
	if _, err := os.Stat(fp); err == nil {
		return fp
	}
	return ""
}

func Load(fn string) (cfg NodeGroup, err error) {
	if fn != "" {
		goto parseAndExit
	}

	// try local dir
	if cwd, err := os.Getwd(); err == nil {
		if fn = load(cwd); fn != "" {
			goto parseAndExit
		}
	}

	// try user config
	if dir, err := ConfigDir(); err != nil {
		return cfg, err
	} else {
		if fn = load(dir); fn != "" {
			goto parseAndExit
		}
	}

	err = errors.New("no config file found")
	return

parseAndExit:
	fn, err = filepath.Abs(fn)
	if err != nil {
		return cfg, err
	}

	logrus.WithField("path", fn).Debug("loading configuration")
	return Parse(fn)
}

func Parse(fn string) (cfg NodeGroup, err error) {
	var b []byte
	b, err = ioutil.ReadFile(fn)
	if err != nil {
		return cfg, err
	}
	switch filepath.Ext(fn) {
	case ".yaml", ".yml":
		b, err = yaml.YAMLToJSONStrict(b)
		if err != nil {
			return cfg, fmt.Errorf("error parsing yaml: %w", err)
		}
		fallthrough
	case ".json":
		err = json.Unmarshal(b, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("error parsing json: %w", err)
		}
	default:
		err = errors.Errorf("format not supported: %s", filepath.Ext(fn))
	}
	return cfg, err
}

func Write(w io.Writer, cfg NodeGroup, format string) (err error) {
	switch format {
	case "yaml", "yml":
		var b []byte
		b, err = yaml.Marshal(cfg)
		if err != nil {
			return
		}
		_, err = w.Write(b)
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		err = enc.Encode(cfg)
	default:
		err = errors.Errorf("format not supported: %s", format)
	}
	return
}
