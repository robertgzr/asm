BUILDX ?= docker buildx

binary:
	$(BUILDX) bake -f build.hcl

image:
	$(BUILDX) bake -f build.hcl image

lint:
	$(BUILDX) bake -f build.hcl lint

debug:
	@CGO_ENABLED=0 go build -tags balena_compat ./cmd/asm
