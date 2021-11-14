package podman

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/buildx/driver"
	dockerclient "github.com/docker/docker/client"
)

func init() {
	driver.Register(&factory{})
}

type factory struct{}

func (*factory) Name() string {
	return "podman"
}

func (*factory) Usage() string {
	return "podman"
}

func (*factory) Priority(_ context.Context, _ dockerclient.APIClient) int {
	return 1 // FIXME
}

func (f *factory) New(ctx context.Context, cfg driver.InitConfig) (driver.Driver, error) {
	d := &Driver{factory: f, InitConfig: cfg}
	for k, v := range cfg.DriverOpts {
		switch {
		case k == "network":
			return nil, fmt.Errorf("option %q not implemented", k)
			// d.netMode = v
			// if v == "host" {
			// 	d.InitConfig.BuildkitFlags = append(d.InitConfig.BuildkitFlags, "--allow-insecure-entitlement=network.host")
			// }
		case k == "image":
			d.image = v
		case k == "cgroup-parent":
			return nil, fmt.Errorf("option %q not implemented", k)
		// d.cgroupParent = v
		case strings.HasPrefix(k, "env."):
			return nil, fmt.Errorf("option %q not implemented", k)
			// envName := strings.TrimPrefix(k, "env.")
			// if envName == "" {
			// 	return nil, errors.Errorf("invalid env option %q, expecting env.FOO=bar", k)
			// }
			// d.env = append(d.env, fmt.Sprintf("%s=%s", envName, v))
		default:
			return nil, fmt.Errorf("invalid driver option %s for podman driver", k)
		}
	}
	return d, nil
}

func (*factory) AllowsInstances() bool {
	return true
}
