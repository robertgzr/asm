package asm

import (
	"context"

	"github.com/docker/buildx/build"
	"github.com/docker/buildx/driver"
	dockercontext "github.com/docker/cli/cli/context"
	"github.com/docker/cli/cli/context/docker"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/robertgzr/asm/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func NewDockerClient(host string, opts map[string]string) (dockerclient.APIClient, error) {
	ep := docker.Endpoint{
		EndpointMeta: docker.EndpointMeta{
			Host: host,
		},
	}

	if opts != nil {
		caPath, haveCA := opts["ca"]
		certPath, haveCert := opts["cert"]
		keyPath, haveKey := opts["key"]
		if haveCA && haveCert && haveKey {
			var err error
			ep.TLSData, err = dockercontext.TLSDataFromFiles(caPath, certPath, keyPath)
			if err != nil {
				return nil, err
			}
			ep.EndpointMeta.SkipTLSVerify = false
			// FIXME remove those to not confuse the docker-container driver
			delete(opts, "ca")
			delete(opts, "cert")
			delete(opts, "key")
		}
	}

	clientOpts, err := ep.ClientOpts()
	if err != nil {
		return nil, err
	}
	logrus.WithField("tls_verify", !ep.EndpointMeta.SkipTLSVerify).Debugf("driver: connecting to %s", host)
	return dockerclient.NewClientWithOpts(clientOpts...)
}

// func NewKubernetesClient(context string) (driver.KubeClientConfig, error) {
// 	// FIXME not sure if this is bad
// 	dockerCli, err := dockercli.NewDockerCli()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return kubernetes.ConfigFromContext(context, dockerCli.ContextStore())
// }

// TODO how can we make this less docker-dependant
func DriversForNodeGroup(ctx context.Context, ng *config.NodeGroup, contextPathHash string) ([]build.DriverInfo, error) {
	eg, _ := errgroup.WithContext(ctx)

	dis := make([]build.DriverInfo, len(ng.Nodes))
	factories := make(map[string]driver.Factory)

	for _, n := range ng.Nodes {
		_, ok := factories[n.Driver]
		if !ok {
			f := driver.GetFactory(n.Driver, false)
			if f == nil {
				return nil, errors.Errorf("failed to find driver %q", f)
			}
			factories[n.Driver] = f
		}
	}

	for i, n := range ng.Nodes {
		func(i int, n config.Node) {
			eg.Go(func() error {
				di := build.DriverInfo{
					Name:     n.Name,
					Platform: n.Platforms,
				}
				defer func() {
					dis[i] = di
				}()

				dockerapi, err := NewDockerClient(n.Endpoint, n.DriverOpts)
				if err != nil {
					di.Err = err
					return nil
				}
				// TODO: replace with dockerclient.WithAPIVersionNegotiation option in clientForEndpoint
				dockerapi.NegotiateAPIVersion(ctx)

				// var kcc driver.KubeClientConfig
				// kcc, err = NewKubernetesClient(n.Endpoint)
				// if err != nil {
				// 	return err
				// }

				d, err := driver.GetDriver(ctx, "asm_buildkit_"+n.Name, factories[n.Driver], dockerapi, nil, nil, n.Flags, n.ConfigFile, n.DriverOpts, n.Platforms, contextPathHash)
				if err != nil {
					logrus.WithField("driver", n.Name).Error(err)
					di.Err = err
					return nil
				}
				di.Driver = d
				return nil
			})
		}(i, n)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return dis, nil
}
