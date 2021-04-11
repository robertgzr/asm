module github.com/robertgzr/asm

go 1.14

require (
	github.com/containerd/containerd v1.4.0
	github.com/docker/buildx v0.4.1
	github.com/docker/cli v0.0.0-20200227165822-2298e6a3fe24
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	sigs.k8s.io/yaml v1.1.0
// gotest.tools v2.2.0+incompatible
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200512144102-f13ba8f2f2fd
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200310163718-4634ce647cf2+incompatible
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)

replace github.com/robertgzr/asm => ./
