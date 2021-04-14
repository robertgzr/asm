package asm

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/buildx/bake"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type targetMap map[string]*bake.Target

func ReadTargets(ctx context.Context, files []bake.File, targets, overrides []string) (targetMap, error) {
	m, err := bake.ReadTargets(ctx, files, targets, overrides)
	if err != nil {
		if !strings.Contains(err.Error(), "unsupported Compose file version") {
			return nil, err
		}
		m = make(targetMap)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name, "docker-compose.yml") {
			v := make(map[string]interface{})
			if err := yaml.Unmarshal(f.Data, &v); err != nil {
				return nil, err
			}
			if verI, ok := v["version"]; ok {
				if ver, ok := verI.(string); ok && strings.HasPrefix(ver, "2") {
					logrus.Warnf("parse: compose version: %q, falling back to legacy parsing", ver)
					c, err := parseCompose(f.Name, v)
					if err != nil {
						return nil, err
					}
					o, err := newOverrides(c, overrides)
					if err != nil {
						return nil, err
					}
					for _, n := range targets {
						for _, n := range c.ResolveGroup(n) {
							t, err := c.ResolveTarget(n, o)
							if err != nil {
								return nil, err
							}
							if t != nil {
								m[n] = t
							}
						}
					}
				}
			}
		}
	}

	return m, nil
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

		balenaCompat(fn, &t)

		if _, ok := sv["image"]; ok {
			itag, ok := sv["image"].(string)
			if !ok {
				return nil, errors.New("parse error: invalid image field")
			}
			t.Tags = []string{itag}
		} else {
			// FIXME ?
			absProjDir, err := filepath.Abs(filepath.Dir(fn))
			if err != nil {
				return nil, fmt.Errorf("parse error: %s", err)
			}
			t.Tags = []string{filepath.Base(absProjDir) + "_" + t.Name}
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
