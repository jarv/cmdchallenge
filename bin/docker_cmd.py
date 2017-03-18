# -*- coding: utf-8 -*-

import docker
from docker.errors import ContainerError, NotFound
from base64 import b64encode
import signal
import re
from os import path
import uuid
from ssl import SSLError
import logging

LOG = logging.getLogger()
DOCKER_TIMEOUT = 8
CMD_TIMEOUT = 4
WORKING_DIR = '/var/challenges'
DOCKER_OPTS = dict(mem_limit='4MB', working_dir=WORKING_DIR,
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

    client = docker.DockerClient(version=docker_version, base_url=docker_base_url, tls=tls_config)
    b64cmd = b64encode(cmd)
    challenge_dir = path.join(WORKING_DIR, challenge['slug'])
    return_token = uuid.uuid4()
    script_name = uuid.uuid4()
    docker_cmd = "cd {challenge_dir} && echo {b64cmd} | base64 -d > /tmp/.{script_name} && timeout {timeout} bash -O globstar /tmp/.{script_name}; echo {return_token}$?".format(
        challenge_dir=challenge_dir,
        b64cmd=b64cmd,
        timeout=CMD_TIMEOUT,
        return_token=return_token,
        script_name=script_name)

    if 'tests' in challenge:
        token = uuid.uuid4()
        test_cmd = ""
        for t in challenge['tests']:
            test_cmd += "if ! {test}; then echo {token}{msg};fi\n".format(
                token=token,
                test=t['test'],
                msg=t['msg'])
        b64testcmd = b64encode(test_cmd)
        docker_cmd += " && cd {challenge_dir} && echo {b64testcmd} | base64 -d > /tmp/test.sh && timeout {timeout} bash -O globstar -e /tmp/test.sh".format(
            challenge_dir=challenge_dir,
            b64testcmd=b64testcmd,
            timeout=CMD_TIMEOUT)

    docker_cmd = "bash -c '{}'".format(docker_cmd)
    return_code = -1
    test_errors = None
    with timeout(seconds=DOCKER_TIMEOUT):
        try:
            LOG.debug("Running `{}` in container".format(docker_cmd))
            output = client.containers.run('cmdline', docker_cmd, **DOCKER_OPTS)
            return_code_match = re.search(r'{}(\d+)'.format(return_token), output)
            if return_code_match is None:
                raise DockerValidationError("Unable to determine return code from command")
            return_code = int(return_code_match.group(1))
            output = re.sub(r'{}\d+'.format(return_token), '', output).rstrip()
            output = re.sub(r'/tmp/.{}: line \d+: (.*)'.format(script_name), r'\1', output)
            if return_code == 124:
                output += "\n** Command timed out after {} seconds **".format(CMD_TIMEOUT)
            if 'tests' in challenge:
                test_errors_matches = re.findall(r'{}(.*)'.format(token), output)
                if test_errors_matches:
                    test_errors = test_errors_matches
                    output = re.sub(r'{}.*'.format(token), '', output, re.M).rstrip()
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
    return output.rstrip(), return_code, test_errors
