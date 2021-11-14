
# development

Check out the [`Makefile`](./Makefile), everything should be in there...
The `binary`, `image` and `lint` targets rely on `docker buildx` (or `asm` if available).

Alternatively docker with buildkit can be used:

```
DOCKER_BUILDKIT=1 docker build --target=binary --output type=local,dest=. .

DOCKER_BUILDKIT=1 docker build --target=image --tag robertgzr/asm:latest .

DOCKER_BUILDKIT=1 docker build --target=lint --output type=tar,dest=/dev/null .
```

If you're all set for Golang development locally, `make debug` should work.


# testing

There are test projects and a sample `asm.yml` under [`./test`](./test).
