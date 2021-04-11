package config

import (
	"encoding/json"
	"os"

	"github.com/docker/buildx/store"
)

type Config struct {
	Debug     bool
	NodeGroup store.NodeGroup
}

func Load(fn string) (cfg Config, err error) {
	f, err := os.Open(fn)
	if err != nil {
		return
	}
	if err = json.NewDecoder(f).Decode(&cfg); err != nil {
		return
	}
	return cfg, nil
}
