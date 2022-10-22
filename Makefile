CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge
PWD=$(shell pwd)
DATE_TS=$(shell date -u +%Y%m%d%H%M%S)
AWS_EXTRA_ARGS=$(shell [ -z $$CI ] && echo "--profile cmdchallenge" || echo "")
DIR_CMDCHALLENGE=$(CURDIR)/cmdchallenge
DIR_SITE=$(CURDIR)/site
DIR_DIST=$(DIR_SITE)/dist
AWS_PROFILE := cmdchallenge

S3_RELEASE_BUCKET_PROD := prod-cmd-release
S3_RELEASE_BUCKET_TESTING := testing-cmd-release

all: upload-testing
prod: upload-prod

.PHONY: upload-testing
upload-testing:
	./bin/upload-cmdchallenge testing

.PHONY: upload-prod
upload-prod:
	./bin/upload-cmdchallenge

.PHONY: build
build:
	./bin/build-cmdchallenge

###################
# CMD Challenge
###################

.PHONY: test
test:
	cd $(DIR_CMDCHALLENGE); go test ./...

.PHONY: update-cmdchallenge
update-cmdchallenge:
	./bin/update-cmdchallenge

.PHONY: build-static
build-static: update-cmdchallenge
	cd $(DIR_SITE); npx vite build

.PHONY: cmdshell
cmdshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDCHALLENGE)/ro_volume",target=/ro_volume $(CI_REGISTRY_IMAGE)/cmd:latest bash

.PHONY: oopsshell
oopsshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDCHALLENGE)/ro_volume",target=/ro_volume $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest bash

.PHONY: clean
clean:
	rm -f $(DIR_CMDCHALLENGE)/ro_volume/ch/*
	rm -f $(DIR_SITE)/challenges.json
