package asm

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/docker/buildx/bake"
	"github.com/pkg/errors"
)

func expandTargets(c *bake.Config, pattern string) ([]string, error) {
	for _, target := range c.Targets {
		if target.Name == pattern {
			return []string{pattern}, nil
		}
	}

	var names []string
	for _, target := range c.Targets {
		ok, err := path.Match(pattern, target.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "could not match targets with '%s'", pattern)
		}
		if ok {
			names = append(names, target.Name)
		}
	}
	if len(names) == 0 {
		return nil, errors.Errorf("could not find any target matching '%s'", pattern)
	}
	return names, nil
}

func newOverrides(c *bake.Config, v []string) (map[string]*bake.Target, error) {
	m := map[string]*bake.Target{}
	for _, v := range v {

		parts := strings.SplitN(v, "=", 2)
		keys := strings.SplitN(parts[0], ".", 3)
		if len(keys) < 2 {
			return nil, errors.Errorf("invalid override key %s, expected target.name", parts[0])
		}

		pattern := keys[0]
		if len(parts) != 2 && keys[1] != "args" {
			return nil, errors.Errorf("invalid override %s, expected target.name=value", v)
		}

		names, err := expandTargets(c, pattern)
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			t, ok := m[name]
			if !ok {
				t = &bake.Target{}
			}

			switch keys[1] {
			case "context":
				t.Context = &parts[1]
			case "dockerfile":
				t.Dockerfile = &parts[1]
			case "args":
				if len(keys) != 3 {
					return nil, errors.Errorf("invalid key %s, args requires name", parts[0])
				}
				if t.Args == nil {
					t.Args = map[string]string{}
				}
				if len(parts) < 2 {
					v, ok := os.LookupEnv(keys[2])
					if ok {
						t.Args[keys[2]] = v
					}
				} else {
					t.Args[keys[2]] = parts[1]
				}
			case "labels":
				if len(keys) != 3 {
					return nil, errors.Errorf("invalid key %s, labels requires name", parts[0])
				}
				if t.Labels == nil {
					t.Labels = map[string]string{}
				}
				t.Labels[keys[2]] = parts[1]
			case "tags":
				t.Tags = append(t.Tags, parts[1])
			case "cache-from":
				t.CacheFrom = append(t.CacheFrom, parts[1])
			case "cache-to":
				t.CacheTo = append(t.CacheTo, parts[1])
			case "target":
				s := parts[1]
				t.Target = &s
			case "secrets":
				t.Secrets = append(t.Secrets, parts[1])
			case "ssh":
				t.SSH = append(t.SSH, parts[1])
			case "platform":
				t.Platforms = append(t.Platforms, parts[1])
			case "output":
				t.Outputs = append(t.Outputs, parts[1])
			case "no-cache":
				noCache, err := strconv.ParseBool(parts[1])
				if err != nil {
					return nil, errors.Errorf("invalid value %s for boolean key no-cache", parts[1])
				}
				t.NoCache = &noCache
			case "pull":
				pull, err := strconv.ParseBool(parts[1])
				if err != nil {
					return nil, errors.Errorf("invalid value %s for boolean key pull", parts[1])
				}
				t.Pull = &pull
			default:
				return nil, errors.Errorf("unknown key: %s", keys[1])
			}
			m[name] = t
		}
	}
	return m, nil
}
