package asm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/buildx/bake"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

func ParseConfig(fn string) (*bake.Config, error) {
	var c *bake.Config

	dt, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	fnl := strings.ToLower(fn)
	if strings.HasSuffix(fnl, ".json") {
		if err := json.Unmarshal(dt, c); err != nil {
			return nil, err
		}
		return c, nil
	}
	if strings.HasSuffix(fnl, "docker-compose.yml") {
		// TODO see if we want to extend this to all compose files
		// strings.HasSuffix(fnl, ".yml") || strings.HasSuffix(".yaml")
		v := make(map[string]interface{})
		if err := yaml.Unmarshal(dt, &v); err != nil {
			return nil, err
		}
		if verI, ok := v["version"]; ok {
			if ver, ok := verI.(string); ok && strings.HasPrefix(ver, "2") {
				logrus.Warnf("parse: compose version: %q, falling back to legacy parsing", ver)
				return parseCompose(fnl, v)
			}
		}
	}

	// handled by upstream
	return bake.ParseFile(dt, fn)
}

func parseCompose(fn string, v map[string]interface{}) (*bake.Config, error) {
	var (
		c              bake.Config
		defaultTargets []string
	)

	svs, ok := v["services"].(map[string]interface{})
	if !ok {
		return nil, errors.New("parse error: invalid services field")
	}

	for svname, svi := range svs {
		var t bake.Target

		sv, ok := svi.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("parse error: invalid service %q field", svname)
		}

		t.Name = svname
		if _, ok := sv["build"]; !ok {
			continue
		}
		if buildDir, ok := sv["build"].(string); ok {
			t.Context = &buildDir
			df := "Dockerfile"
			t.Dockerfile = &df
		} else {
			logrus.Warnf("%#+v", sv["build"])
			b, ok := sv["build"].(map[string]interface{})
			if !ok {
				return nil, errors.New("parse error: invalid build field")
			}
			if v, ok := b["context"].(string); ok {
				t.Context = &v
			}
			if v, ok := b["dockerfile"].(string); ok {
				t.Dockerfile = &v
			}
			if v, ok := b["args"].(map[string]string); ok {
				t.Args = v
			}
		}

		// FIXME: find a better spot for this
		balenaCompat(fn, &t)

		if _, ok := sv["image"]; ok {
			itag, ok := sv["image"].(string)
			if !ok {
				return nil, errors.New("parse error: invalid image field")
			}
			t.Tags = append(t.Tags, itag)
		}
		c.Targets = append(c.Targets, &t)
		defaultTargets = append(defaultTargets, t.Name)
	}

	c.Groups = []*bake.Group{{
		Name:    "default",
		Targets: defaultTargets,
	}}

	return &c, nil
}
