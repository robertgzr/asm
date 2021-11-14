BINARY ?= asm
export ASM_BINARY = $(abspath ${BINARY})

# --- CONTAINER

buildx = $(if $(shell test -x ${ASM_BINARY} && echo 1),${ASM_BINARY},docker buildx)

binary:
	$(call buildx) bake -f build.hcl

image:
	$(call buildx) bake -f build.hcl image

lint:
	$(call buildx) bake -f build.hcl lint

# --- LOCAL

debug: BUILDTAGS ?= osusergo netgo static_build balena podman
debug:
	CGO_ENABLED=0 go build -tags "${BUILDTAGS}" -o ${ASM_BINARY} ./cmd/asm

install: PREFIX ?= /usr/local
install:
	install -t $(PREFIX)/bin/ asm

# ---- TEST

test: debug
	$(MAKE) -C $@ all

test-balena: debug
	$(MAKE) -C test balena

test-default: debug
	$(MAKE) -C test default
