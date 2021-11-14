# asm

... is [bake](https://github.com/docker/buildx/blob/master/docs/reference/buildx_bake.md)

* compatible with docker-compose version 2
* compatible with balena projects [*](#balena)
* easy driver configuration

* _for now_ without the kubernetes driver
* with a [podman](https://podman.io/) driver

## usage
### via cli
```
asm gen docker > asm.yml
asm bake -f compose.yaml
```
### via container image
```
docker run --rm -it \
	--mount type=bind,src=${PWD}/asm.yml:/run/asm/asm.yml:ro \
	--mount type=bind,src=${PWD}/compose.yaml:/run/asm/compose.yaml:ro \
	robertgzr/asm:latest \
		bake -f ompose.yaml
```

## podman

The podman driver, relies on the binary being available in your `PATH`.
When using rootless podman, the driver needs to have the `rootless: true` option.
See https://github.com/moby/buildkit/blob/master/docs/rootless.md

## balena

Supports [`Dockerfile.template` handling][balena-template], and [build time secrets/variables][balena-secret].

Template variable values can be set via environment:

Template variable   | Environment variable
--------------------|---------------------
BALENA_MACHINE_NAME | ASM_BALENA_MACHINE_NAME
BALENA_ARCH         | ASM_BALENA_ARCH
BALENA_APP_NAME     | ASM_BALENA_APP_NAME
BALENA_RELEASE_HASH | ASM_BALENA_RELEASE_HASH

**For secrets, there is a caveat:**  
Populating the build image with the secret relies on balenaEngine, `asm` uses buildkit instead.
As such the [buildkit syntax][buildkit-secret] needs to be used, rendering this incompatible
with balena.

```Dockerfile
RUN --mount=type=secret,id=my-secret.txt \
	test -f /run/secrets/my-secret.txt
```

[balena-template]: https://www.balena.io/docs/learn/deploy/deployment/#template-files
[balena-secret]: https://www.balena.io/docs/learn/deploy/deployment/#build-time-secrets-and-variables
[buildkit-secret]: https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md#run---mounttypesecret
