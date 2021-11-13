# syntax=docker/dockerfile:1.2

ARG BUILDTAGS='netgo osusergo static_build balena_compat'
ARG GO_VERSION=1.16

FROM --platform=$BUILDPLATFORM tonistiigi/xx:golang AS xgo
# base stage
FROM --platform=$BUILDPLATFORM docker.io/library/golang:$GO_VERSION-alpine AS base
RUN apk add -U --no-cache ca-certificates

# development (build) stage
FROM base AS dev
ARG BUILDPLATFORM
RUN apk add -U --no-cache file git
# add cross compile helpers
COPY --from=xgo / /
WORKDIR /src
ENTRYPOINT [ "sh" ]

FROM dev AS version
RUN --mount=target=. \
    set -ex; \
    PKG=github.com/robertgzr/asm; \
    VERSION=$(git describe --match 'v[0-9]*' --dirty='+dirty' --always --tags); \
    REVISION=$(git rev-parse HEAD)$(if ! git diff --no-ext-diff --quiet --exit-code; then echo '+dirty'; fi); \
    echo "-X ${PKG}/version.Version=${VERSION} -X ${PKG}/version.Revision=${REVISION} -X ${PKG}/version.Package=${PKG}" | \
      tee /tmp/.ldflags; \
    echo "${VERSION}" | tee /tmp/.version;

FROM dev AS lint
ARG BUILDTAGS
ARG TARGETPLATFORM
# install golangci-lint
RUN wget -q https://github.com/golangci/golangci-lint/releases/download/v1.24.0/golangci-lint-1.24.0-$(echo $BUILDPLATFORM | sed 's/\//-/g').tar.gz -O - | tar -xzf - -C /usr/local/bin/ --strip-components=1 && chmod +x /usr/local/bin/golangci-lint
RUN --mount=target=. \
    --mount=target=/go/pkg,type=cache \
    --mount=target=/root/.cache,type=cache \
    CGO_ENABLED=0 \
    golangci-lint run --build-tags "${BUILDTAGS}" ./... || true

FROM dev AS gobuild
ARG BUILDTAGS
ARG TARGETPLATFORM
RUN --mount=target=. \
    --mount=target=/go/pkg,type=cache \
    --mount=target=/root/.cache,type=cache \
    --mount=target=/tmp/.ldflags,source=/tmp/.ldflags,from=version \
    set -ex; \
    go build \
      -tags "${BUILDTAGS}" \
      -ldflags "$(cat /tmp/.ldflags)" \
      -o /out/asm ./cmd/asm; \
    file /out/asm | grep "statically linked"

# binaries
FROM scratch AS binary
COPY --from=gobuild /out/* /

# unit tests
# FROM dev AS test
# ARG BUILDTAGS
# RUN --mount=target=. \
#     --mount=target=/go/pkg,type=cache \
#     --mount=target=/root/.cache,type=cache \
#       hack/test

# # integration tests
# FROM dev AS test-integration
# COPY --from=final /* /usr/local/bin/
# ARG BUILDTAGS
# RUN --mount=target=. \
#     --mount=target=/go/pkg,type=cache \
#     --mount=target=/root/.cache,type=cache \
#       hack/test-integration

# container image
FROM base AS image
COPY --from=gobuild /out/* /usr/local/bin/
WORKDIR /run/asm

# generate a default config
ARG TARGETPLATFORM
RUN asm gen docker -H unix://var/run/docker.sock --platform ${TARGETPLATFORM} >config.json

ENTRYPOINT [ "/usr/local/bin/asm" ]

# vim: ft=dockerfile
