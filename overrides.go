package asm

import (
	"os"
	"path"
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

func newOverrides(c *bake.Config, v []string) (map[string]map[string]bake.Override, error) {
	m := map[string]map[string]bake.Override{}
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

		kk := strings.SplitN(parts[0], ".", 2)

		for _, name := range names {
			t, ok := m[name]
			if !ok {
				t = map[string]bake.Override{}
				m[name] = t
			}

			o := t[kk[1]]

			switch keys[1] {
			case "output", "cache-to", "cache-from", "tags", "platform", "secrets", "ssh":
				if len(parts) == 2 {
					o.ArrValue = append(o.ArrValue, parts[1])
				}
			case "args":
				if len(keys) != 3 {
					return nil, errors.Errorf("invalid key %s, args requires name", parts[0])
				}
				if len(parts) < 2 {
					v, ok := os.LookupEnv(keys[2])
					if !ok {
						continue
					}
					o.Value = v
				}
				fallthrough
			default:
				if len(parts) == 2 {
					o.Value = parts[1]
				}
			}
			t[kk[1]] = o
		}
	}
	return m, nil
}
