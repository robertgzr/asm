module github.com/robertgzr/asm

go 1.14

require (
	github.com/containerd/containerd v1.5.0-beta.4
	github.com/docker/buildx v0.5.1
	github.com/docker/cli v20.10.5+incompatible
	github.com/docker/docker v20.10.5+incompatible
	github.com/moby/buildkit v0.8.1-0.20201205083753-0af7b1b9c693
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/robertgzr/asm => ./
