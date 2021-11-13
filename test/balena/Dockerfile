ARG BALENA_MACHINE_NAME
FROM balenalib/${BALENA_MACHINE_NAME}-alpine:run

ARG BALENA_ARCH
ARG BALENA_SERVICE_NAME
RUN test -n "${BALENA_ARCH}" && test -n "${BALENA_SERVICE_NAME}"

RUN --mount=type=secret,id=my-secret.txt \
	test -f /run/secrets/my-secret.txt

ARG TEST
RUN test -n "${TEST}" && test "${TEST}" = "TEST"

EXPOSE 8080
