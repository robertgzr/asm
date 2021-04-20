package asm

import (
	"context"

	"github.com/docker/buildx/build"
	"github.com/docker/buildx/driver"
	"github.com/docker/buildx/store"
	"github.com/docker/cli/cli/context/docker"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func NewDockerClient(host string) (dockerclient.APIClient, error) {
	ep := docker.Endpoint{
		EndpointMeta: docker.EndpointMeta{
			Host: host,
		},
	}
	clientOpts, err := ep.ClientOpts()
	if err != nil {
		return nil, err
	}
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

// TODO can we support custom drivers here?
func DriversForNodeGroup(ctx context.Context, ng *store.NodeGroup, contextPathHash string) ([]build.DriverInfo, error) {
	eg, _ := errgroup.WithContext(ctx)

	dis := make([]build.DriverInfo, len(ng.Nodes))

	var f driver.Factory
	if ng.Driver != "" {
		f = driver.GetFactory(ng.Driver, false)
		if f == nil {
			return nil, errors.Errorf("failed to find driver %q", f)
		}
	} else {
		dockerapi, err := NewDockerClient(ng.Nodes[0].Endpoint)
		if err != nil {
			return nil, err
		}
		f, err = driver.GetDefaultFactory(ctx, dockerapi, false)
		if err != nil {
			return nil, err
		}
		ng.Driver = f.Name()
	}

	for i, n := range ng.Nodes {
		func(i int, n store.Node) {
			eg.Go(func() error {
				di := build.DriverInfo{
					Name:     n.Name,
					Platform: n.Platforms,
				}
				defer func() {
					dis[i] = di
				}()

				dockerapi, err := NewDockerClient(n.Endpoint)
				if err != nil {
					di.Err = err
					return nil
				}
				// TODO: replace the following line with dockerclient.WithAPIVersionNegotiation option in clientForEndpoint
				dockerapi.NegotiateAPIVersion(ctx)

				// var kcc driver.KubeClientConfig
				// kcc, err = NewKubernetesClient(n.Endpoint)
				// if err != nil {
				// 	return err
				// }

				d, err := driver.GetDriver(ctx, "asm_buildkit_"+n.Name, f, dockerapi, nil, nil, n.Flags, n.ConfigFile, n.DriverOpts, n.Platforms, contextPathHash)
				if err != nil {
					di.Err = err
					logrus.WithField("driver", n.Name).Error(di.Err)
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
