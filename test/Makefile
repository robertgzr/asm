ASM_OPTS ?=

.PHONY: all
all: default balena

.PHONY: default
default:
	cd $@ && ${ASM_BINARY} ${ASM_OPTS} bake -f docker-compose.yml

.PHONY: balena
balena:
	cd $@ && ${ASM_BINARY} ${ASM_OPTS} bake -f docker-compose.yml
