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

Start the backend the `-dev` option uses an in-memory db. Without it, an sqlite db will be created `cmdchallenge/db.sql`.


**Backend:**

```
make build # builds the docker images for the runner
cd cmdchallenge
go run cmd/serve/serve.go -dev
```

**Frontend:**

```
cd site
npx vite
```

## Misc

**Test a single command:**

```
curl  http://localhost:8181/c/r -X POST -F slug=hello_world -F cmd="echo hello world"
```

**Fetch solutions:**

```
curl http://localhost:8181/c/s?slug=hello_world
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
