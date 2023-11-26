# syntax = docker/dockerfile:1-experimental

ARG BUILD_PLATFORM
FROM --platform=${BUILD_PLATFORM} golang:1.21 as runcmd-builder
WORKDIR /app
COPY cmdchallenge .
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags "-w" -o runcmd ./cmd/runcmd/runcmd.go

FROM node:20.2.0-bullseye-slim as site-builder
WORKDIR /app
COPY site .
RUN npm install && \
  rm -rf ./dist && \
  npx vite build

FROM --platform=${BUILD_PLATFORM} debian:bookworm-slim
COPY --from=runcmd-builder /app/runcmd /app/runcmd
COPY --from=site-builder /app/dist /app/dist

RUN apt-get update && \
  groupadd --force -g 500 docker && \
  apt-get install -y docker.io && \
  useradd -u 510 -G docker -r cmd && \
  chown cmd /app

CMD ["/app/runcmd"]
