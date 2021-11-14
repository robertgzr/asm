module github.com/robertgzr/asm

go 1.16

require (
	github.com/containerd/containerd v1.5.5
	github.com/docker/docker v20.10.7+incompatible
	github.com/opencontainers/image-spec v1.0.2-0.20210819154149-5ad6f50d6283
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	sigs.k8s.io/yaml v1.2.0
	go.opentelemetry.io/otel v1.0.0-RC1
)

// NOTE: make sure these are in sync with buildx
require (
	github.com/docker/buildx v0.7.0
	github.com/moby/buildkit v0.9.1-0.20211019185819-8778943ac3da
// github.com/moby/buildkit v0.9.1
)

replace (
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20210702143511-f782d1355eff+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.3-0.20210817025855-ba2adeebdb8d+incompatible
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => github.com/tonistiigi/opentelemetry-go-contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.0.0-20210714055410-d010b05b4939
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace => github.com/tonistiigi/opentelemetry-go-contrib/instrumentation/net/http/httptrace/otelhttptrace v0.0.0-20210714055410-d010b05b4939
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => github.com/tonistiigi/opentelemetry-go-contrib/instrumentation/net/http/otelhttp v0.0.0-20210714055410-d010b05b4939
)

replace github.com/robertgzr/asm => ./
