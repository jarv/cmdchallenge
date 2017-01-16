# -*- coding: utf-8 -*-

import docker
from docker.errors import ContainerError, NotFound
from base64 import b64encode
import signal
import re

DOCKER_TIMEOUT = 8
CMD_TIMEOUT = 4
WORKING_DIR = '/var/challenges'
DOCKER_OPTS = dict(mem_limit='4MB', working_dir=WORKING_DIR,
                   network_mode=None, network_disabled=True)


class TimeoutError(Exception):
    pass


class timeout:
    def __init__(self, seconds=1, error_message='Timeout'):
        self.seconds = seconds
        self.error_message = error_message

    def handle_timeout(self, signum, frame):
        raise TimeoutError(self.error_message)

    def __enter__(self):
        signal.signal(signal.SIGALRM, self.handle_timeout)
        signal.alarm(self.seconds)

    def __exit__(self, type, value, traceback):
        signal.alarm(0)


def output_from_cmd(cmd, challenge, docker_version=None, docker_base_url=None):
    client = docker.DockerClient(version=docker_version, base_url=docker_base_url)
    b64cmd = b64encode(cmd)
    docker_cmd = "bash -c 'cd {}; echo {} | base64 -d > /tmp/script.sh; timeout {} bash -ex /tmp/script.sh'".format(
        challenge['slug'], b64cmd, CMD_TIMEOUT)
    return_code = 1
    with timeout(seconds=DOCKER_TIMEOUT):
        try:
            output = client.containers.run('cmdline', docker_cmd, **DOCKER_OPTS)
            return_code = 0
        except ContainerError as e:
            output = re.sub(r'/tmp/script.sh: line \d+: (.*)', r'\1', e.stderr)
            return_code = e.exit_status
            if return_code == 124:
                output += "\n** Command timed out after {} seconds **".format(CMD_TIMEOUT)
        except NotFound as e:
            output = e.explanation
        except TimeoutError as e:
            output = "Command timed out"
    return output.rstrip(), return_code
