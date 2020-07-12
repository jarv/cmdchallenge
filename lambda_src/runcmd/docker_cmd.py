import docker
from docker.errors import ContainerError, NotFound, APIError
from requests import ConnectionError
from base64 import b64encode
import signal
from os import environ
from os.path import dirname, realpath, join, isdir
import json
from ssl import SSLError
import logging

LOG = logging.getLogger()
DOCKER_TIMEOUT = 8
BASE_WORKING_DIR = "/var/challenges"

IMAGE_TAG = "latest" if environ.get('IS_PROD') == "yes" else "testing"
REGISTRY_IMAGE = f"registry.gitlab.com/jarv/cmdchallenge/cmd:{IMAGE_TAG}"

dir_path = dirname(realpath(__file__))
dir_cmdchallenge = join(dir_path, "../../cmdchallenge")

if environ.get('LAMBDA_RUNTIME_DIR'):
    volume_dir = "/var/ro_volume"
else:
    volume_dir = join(dir_cmdchallenge, "ro_volume")
    if not isdir(volume_dir):
        LOG.error(f"{volume_dir} is not a directory!")
        raise (Exception("System error"))

DOCKER_OPTS = dict(
    mem_limit="10MB",
    volumes={volume_dir: {"bind": "/ro_volume", "mode": "ro"}},
    network_mode=None,
    network_disabled=True,
    remove=True,
    stderr=True,
    detach=False,
)


class DockerValidationError(Exception):
    pass


class CommandTimeoutError(Exception):
    pass


class timeout:
    def __init__(self, seconds=1, error_message="Timeout"):
        self.seconds = seconds
        self.error_message = error_message

    def handle_timeout(self, signum, frame):
        raise CommandTimeoutError(self.error_message)

    def __enter__(self):
        signal.signal(signal.SIGALRM, self.handle_timeout)
        signal.alarm(self.seconds)

    def __exit__(self, type, value, traceback):
        signal.alarm(0)


def output_from_cmd(
    cmd, challenge, docker_version=None, docker_base_url=None, tls_settings=None
):
    if tls_settings:
        tls_config = docker.tls.TLSConfig(**tls_settings)
    else:
        tls_config = None

    if docker_base_url:
        LOG.debug(
            f"Using GitLab CI docker configuration: {docker_version} / {docker_base_url} / {tls_settings}"
        )
        client = docker.DockerClient(
            version=docker_version, base_url=docker_base_url, tls=tls_config
        )
    else:
        client = docker.from_env()

    b64cmd = b64encode(cmd.encode("utf-8"))
    challenge_dir = join(BASE_WORKING_DIR, challenge["slug"])
    docker_cmd = f'runcmd --slug {challenge["slug"]} {b64cmd.decode("utf-8")}'

    with timeout(seconds=DOCKER_TIMEOUT):
        try:
            LOG.debug(f"Running `{docker_cmd}` in container")

            output = client.containers.run(
                REGISTRY_IMAGE, docker_cmd, working_dir=challenge_dir, **DOCKER_OPTS
            ).decode("utf-8")
        except SSLError:
            LOG.exception(f"SSL validation error connecting to {docker_base_url}")
            raise DockerValidationError("SSL Error")
        except ContainerError as e:
            LOG.exception("Container error")
            raise DockerValidationError(
                f"There was a problem executing the command, return code: {e.exit_status}"
            )
        except NotFound as e:
            LOG.exception("NotFound error")
            raise DockerValidationError(e.explanation)
        except CommandTimeoutError:
            LOG.exception("CommandTimeout error")
            raise DockerValidationError("Command timed out")
        except APIError:
            LOG.exception("Docker API error")
            raise DockerValidationError("Docker API error")
        except ConnectionError:
            LOG.exception("Docker ConnectionError")
            raise DockerValidationError("Docker connection error")
        try:
            output_json = json.loads(output)
        except ValueError:
            LOG.exception("JSON decode error")
            raise DockerValidationError("Command failure")
    if "Error" in output_json:
        LOG.error("Command execution error: {}".format(output_json["Error"]))
        raise DockerValidationError("Command execution error")
    return output_json
