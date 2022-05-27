# CMD Challenge

This repository contains the code for the site [cmdchallenge.com](https://cmdchallenge.com)

[Read more about cmdchallenge](https://jarv.org/tags/cmdchallenge/)

## Installation

- [Install Docker](https://docs.docker.com/get-docker/)
- [Install `asdf`](http://asdf-vm.com/guide/getting-started.html#_1-install-dependencies)
- `asdf install`
- `docker pull registry.gitlab.com/jarv/cmdchallenge/cmd`
- `docker pull registry.gitlab.com/jarv/cmdchallenge/cmd-no-bin`

## Testing

- `make test`

## Local development

### Backend

Start the backend which will also initialize a new sqlite database in the `cmdchallenge/` directory.

If you want to use the test in-memory database use the `-dev` flag.

```
make build-image-cmd # builds the docker images for the runner
go run cmdchallenge/cmd/serve.go
```

Test a single command:

```
curl  http://localhost:8181/c/r -X POST -F slug=hello_world -F cmd="echo hello world"
```

Fetch solutions:

```
curl http://localhost:8181/c/s?slug=hello_world
```

### Frontend

```
cd site
npx vite
```

## CI vars

The following CI vars are necessary to run the full pipeline

- `AWS_ACCESS_KEY_ID`: Access key for AWS
- `AWS_SECRET_ACCESS_KEY`: Secret key for AWS
- `STATE_S3_BUCKET`: where to store Terraform state
- `STATE_S3_KEY`: key for storing state
- `STATE_S3_REGION`: region for deployment
- `SSH_PRIVATE_KEY`: Private SSH key for the remote Docker machine
- `SSH_PUBLIC_KEY` : Public SSH key for the remote Docker machine

## Bugs / Suggestions

- Open [a GitLab issue](https://gitlab.com/jarv/cmdchallenge/-/issues).
