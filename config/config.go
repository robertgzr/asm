package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/buildx/store"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

type NodeGroup struct {
	Nodes []Node
}

type Node struct {
	store.Node
	Driver string
}

func load(dir string) string {
	var fp string
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
		goto exit
	}

	// try local dir
	if fn = load("."); fn != "" {
		goto exit
	}

	// try xdg
	if xdg, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		if fn = load(filepath.Join(xdg, "asm")); fn != "" {
			goto exit
		}
	}

	err = errors.New("no config file found")
	return

exit:
	return Parse(fn)
}

func Parse(fn string) (cfg NodeGroup, err error) {
	var f io.Reader
	f, err = os.Open(fn)
	if err != nil {
		return
	}
	switch filepath.Ext(fn) {
	case ".yaml", ".yml":
		err = yaml.NewDecoder(f).Decode(&cfg)
	case ".json":
		err = json.NewDecoder(f).Decode(&cfg)
	default:
		err = errors.Errorf("format not supported: %s", filepath.Ext(fn))
	}
	return
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
