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

func ReadTargets(ctx context.Context, files []bake.File, targets, overrides []string, defaults map[string]string) (targetMap, error) {
	m, _, err := bake.ReadTargets(ctx, files, targets, overrides, defaults)
	if err != nil {
		if !strings.Contains(err.Error(), "unsupported Compose file version") {
			return nil, err
		}
		m = make(targetMap)
	}

	// handle v2 compose-file
	for _, f := range files {
		if strings.HasSuffix(f.Name, "docker-compose.yml") {
			v := make(map[string]interface{})
			if err := yaml.Unmarshal(f.Data, &v, yaml.DisallowUnknownFields); err != nil {
				return nil, err
			}
			if verI, ok := v["version"]; ok {
				if ver, ok := verI.(string); ok && strings.HasPrefix(ver, "2") {
					logrus.Infof("compose version: %q, falling back to legacy parsing", ver)
					c, err := parseCompose(f.Name, v)
					if err != nil {
						return nil, err
					}
					o, err := newOverrides(c, overrides)
					if err != nil {
						return nil, err
					}
					for _, groupName := range targets {
						for _, targetName := range c.ResolveGroup(groupName) {
							t, err := c.ResolveTarget(targetName, o)
							if err != nil {
								return nil, err
							}

							// NOTE: otherwise we can't tell if the field was undefined
							for _, parsedTarget := range c.Targets {
								if parsedTarget.Name == targetName {
									if parsedTarget.Dockerfile == nil {
										t.Dockerfile = nil
									}
								}
							}

							if t != nil {
								m[targetName] = t
							}
						}
					}
				}
			}
		}
	}

	if err := parseBalena(m, files); err != nil {
		return nil, fmt.Errorf("balena compatability layer failed: %w", err)
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
			context := filepath.Join(filepath.Dir(fn), buildDir)
			t.Context = &context
		} else {
			b, ok := sv["build"].(map[string]interface{})
			if !ok {
				return nil, errors.New("parse error: invalid build field")
			}
			if v, ok := b["context"].(string); ok {
				context := filepath.Join(filepath.Dir(fn), v)
				t.Context = &context
			}
			if v, ok := b["dockerfile"].(string); ok {
				t.Dockerfile = &v
			}
			if v, ok := b["args"].(map[string]string); ok {
				t.Args = v
			}
		}

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
