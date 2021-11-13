BUILDX ?= docker buildx
BINARY ?= asm

binary:
	$(BUILDX) bake -f build.hcl

image:
	$(BUILDX) bake -f build.hcl image

lint:
	$(BUILDX) bake -f build.hcl lint

# ---

export ASM_BINARY = $(abspath ${BINARY})

debug: BUILDTAGS ?= balena osusergo netgo static_build
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
