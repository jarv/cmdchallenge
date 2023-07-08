# CMD Challenge

This repository contains the code for the site [cmdchallenge.com](https://cmdchallenge.com)

[Read more about cmdchallenge](https://jarv.org/tags/cmdchallenge/)

## Installation

- [Install Docker](https://docs.docker.com/get-docker/)
- [Install rtx](https://github.com/jdxcode/rtx#quickstart)
- `rtx install`
- `docker-compose build` (for M1 macs run `BUILD_PLATFORM=linux/arm64 docker-compose build`)
- `docker-compose up runcmd --remove-orphans` (for M1 macs run `BUILD_PLATFORM=linux/arm64 docker-compose up runcmd --remove-orphans`)

## Testing

- `cd cmdchallenge && go test ./...`

## Local development

### Static assets

```
cd site
npm install
npx vite build
```

### Run the server

```
cd cmdchallenge
# Start the backend the `-dev` option uses an in-memory db. Without it, a path to a sqlite db will be needed
go run cmd/runcmd/runcmd.go -dev -staticDistDir=../site/dist
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

## Bugs / Suggestions

- Open [a GitHub issue](https://github.com/jarv/cmdchallenge/-/issues).
