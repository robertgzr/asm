ASM_OPTS ?= --config=${PWD}/asm.yml

.PHONY: all
all: default balena

.PHONY: default
default:
	cd $@ && ${ASM_BINARY} ${ASM_OPTS} bake -f docker-compose.yml ${BAKE_OPTS}

.PHONY: balena
balena:
	cd $@ && ${ASM_BINARY} ${ASM_OPTS} bake -f docker-compose.yml ${BAKE_OPTS}
