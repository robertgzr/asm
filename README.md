# asm

... is [bake](https://github.com/docker/buildx/blob/master/docs/reference/buildx_bake.md)

* compatible with docker-compose version 2
* compatible with balena projects
* easy driver configuration

* _for now_ without the kubernetes driver

## via cli

```
asm gen docker > asm.yml
asm bake -f compose.yaml
```

## via container image

```
docker run --rm -it \
	--mount type=bind,src=${PWD}/asm.yml:/run/asm/asm.yml:ro \
	--mount type=bind,src=${PWD}/compose.yaml:/run/asm/compose.yaml:ro \
	robertgzr/asm:latest \
		bake -f ompose.yaml
```


## building

```
make binary
make image
```

or

```
DOCKER_BUILDKIT=1 docker build --target=binary --output type=local,dest=. .
DOCKER_BUILDKIT=1 docker build --target=image --tag robertgzr/asm:latest .
```
