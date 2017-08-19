# -*- coding: utf-8 -*-

import docker
from docker.errors import ContainerError, NotFound, APIError
from docker.utils import kwargs_from_env
from requests import ConnectionError
from base64 import b64encode
import signal
from os import path
from os import environ
import json
from ssl import SSLError
import logging

LOG = logging.getLogger()
DOCKER_TIMEOUT = 8
BASE_WORKING_DIR = '/var/challenges'
SCRIPT_DIR = path.abspath(path.dirname(__file__))
if path.isdir('/var/ro_volume'):
    volume_dir = '/var/ro_volume'
else:
    volume_dir = path.join(SCRIPT_DIR, '..', 'ro_volume')
DOCKER_OPTS = dict(mem_limit='4MB',
                   volumes={volume_dir: {'bind': '/ro_volume', 'mode': 'ro'}},
                   network_mode=None, network_disabled=True,
                   remove=True, stderr=True, stdout=True)


class DockerValidationError(Exception):
    pass


class CommandTimeoutError(Exception):
    pass


class timeout:
    def __init__(self, seconds=1, error_message='Timeout'):
        self.seconds = seconds
        self.error_message = error_message

    def handle_timeout(self, signum, frame):
        raise CommandTimeoutError(self.error_message)

    def __enter__(self):
        signal.signal(signal.SIGALRM, self.handle_timeout)
        signal.alarm(self.seconds)

    def __exit__(self, type, value, traceback):
        signal.alarm(0)


def output_from_cmd(cmd, challenge, docker_version=None, docker_base_url=None, tls_settings=None):
    if tls_settings:
        tls_config = docker.tls.TLSConfig(**tls_settings)
    else:
        tls_config = None

    if environ.get('DOCKER_MACHINE_NAME') is None:
        client = docker.DockerClient(version=docker_version, base_url=docker_base_url, tls=tls_config)
    else:
        client = docker.DockerClient(**kwargs_from_env(assert_hostname=False))

    b64cmd = b64encode(cmd)
    challenge_dir = path.join(BASE_WORKING_DIR, challenge['slug'])
    docker_cmd = "/ro_volume/runcmd -slug {slug} {b64cmd}".format(
        slug=challenge['slug'],
        b64cmd=b64cmd)
    with timeout(seconds=DOCKER_TIMEOUT):
        try:
            LOG.warn("Running `{}` in container".format(docker_cmd))
            output = client.containers.run('registry.gitlab.com/jarv/cmdchallenge', docker_cmd, working_dir=challenge_dir, **DOCKER_OPTS)
        except SSLError as e:
            LOG.exception("SSL validation error connecting to {}".format(docker_base_url))
            raise DockerValidationError("SSL Error")
        except ContainerError as e:
            LOG.exception("Container error")
            raise DockerValidationError("There was a problem executing the command, return code: {}".format(e.exit_status))
        except NotFound as e:
            LOG.exception("NotFound error")
            raise DockerValidationError(e.explanation)
        except CommandTimeoutError as e:
            LOG.exception("CommandTimeout error")
            raise DockerValidationError("Command timed out")
        except APIError as e:
            LOG.exception("Docker API error")
            raise DockerValidationError("Docker API error")
        except ConnectionError as e:
            LOG.exception("Docker ConnectionError")
            raise DockerValidationError("Docker connection error")
        try:
            output_json = json.loads(output)
        except ValueError as e:
            LOG.exception("JSON decode error")
            raise DockerValidationError("Command failure")
    if 'Error' in output_json:
        LOG.error("Command execution error: {}".format(output_json['Error']))
        raise DockerValidationError("Command execution error")
    return output_json
