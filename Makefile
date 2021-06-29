BUILDX ?= docker buildx

binary:
	$(BUILDX) bake -f build.hcl

image:
	$(BUILDX) bake -f build.hcl image

lint:
	$(BUILDX) bake -f build.hcl lint

debug: BUILDTAGS ?= balena_compat osusergo netgo static_build
debug:
	CGO_ENABLED=0 go build -tags "$(BUILDTAGS)" ./cmd/asm

install: PREFIX ?= /usr/local
install:
	install -t $(PREFIX)/bin/ asm
