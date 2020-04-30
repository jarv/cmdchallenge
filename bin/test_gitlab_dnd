#!/usr/bin/env python

import docker

DOCKER_OPTS = dict(
    mem_limit="10MB",
    network_mode=None,
    network_disabled=True,
    remove=True,
    stderr=True,
    detach=False,
)

client = docker.from_env()

print(
    client.containers.run("alpine", "echo hello world", **DOCKER_OPTS).decode("utf-8")
)
