package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/buildx/store"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

type Config = store.NodeGroup

func Load(fn string) (cfg Config, err error) {
	if fn != "" {
		goto exit
	}

	if fi, err := os.Stat("asm.yml"); err == nil {
		fn = fi.Name()
		goto exit
	}
	if fi, err := os.Stat("asm.yaml"); err == nil {
		fn = fi.Name()
		goto exit
	}
	if fi, err := os.Stat("asm.json"); err == nil {
		fn = fi.Name()
		goto exit
	}
	err = errors.New("no config file found")
	return

exit:
	return Parse(fn)
}

func Parse(fn string) (cfg Config, err error) {
	var f io.Reader
	f, err = os.Open(fn)
	if err != nil {
		return
	}
	switch filepath.Ext(fn) {
	case ".yaml", ".yml":
		var b []byte
		b, err = ioutil.ReadAll(f)
		if err != nil {
			return
		}
		err = yaml.Unmarshal(b, &cfg)
	case ".json":
		err = json.NewDecoder(f).Decode(&cfg)
	default:
		err = errors.Errorf("format not supported: %s", filepath.Ext(fn))
	}
	return
}

func Write(w io.Writer, cfg Config, format string) (err error) {
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
