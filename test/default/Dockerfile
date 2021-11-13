FROM balenalib/amd64-alpine:run as base
EXPOSE 8080

FROM base AS stage1
RUN date > /build.timestamp

FROM base
