BINARY ?= asm

buildx = $(if $(shell test -x ${BINARY} && echo 1),${BINARY},docker buildx)

binary:
	$(call buildx) bake -f build.hcl

image:
	$(call buildx) bake -f build.hcl image

lint:
	$(call buildx) bake -f build.hcl lint

# ---

export ASM_BINARY = $(abspath ${BINARY})

debug: BUILDTAGS ?= osusergo netgo static_build balena podman
debug:
	CGO_ENABLED=0 go build -tags "${BUILDTAGS}" -o ${ASM_BINARY} ./cmd/asm

install: PREFIX ?= /usr/local
install:
	install -t $(PREFIX)/bin/ asm

# ----

test: debug
	$(MAKE) -C $@ all

test-balena: debug
	$(MAKE) -C test balena

test-default: debug
	$(MAKE) -C test default
