CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge
PWD=$(shell pwd)
DATE_TS=$(shell date -u +%Y%m%d%H%M%S)
AWS_EXTRA_ARGS=$(shell [ -z $$CI ] && echo "--profile cmdchallenge" || echo "")
DIR_CMDRUNNER=$(CURDIR)/cmdrunner
DIR_CMDCHALLENGE=$(CURDIR)/cmdchallenge
DIR_SITE=$(CURDIR)/site
DIR_DIST=$(DIR_SITE)/dist
AWS_PROFILE := cmdchallenge

S3_RELEASE_BUCKET_PROD := prod-cmd-release
S3_RELEASE_BUCKET_TESTING := testing-cmd-release

all: upload-testing push-image-cmd-testing
prod: upload-prod push-image-cmd-prod

.PHONY: upload-testing
upload-testing: update-challenges build build-static
	aws $(AWS_EXTRA_ARGS) s3 cp cmdchallenge/serve s3://$(S3_RELEASE_BUCKET_TESTING)/serve
	tar zcf /tmp/ro_volume.tar.gz -C cmdchallenge ro_volume/
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/ro_volume.tar.gz s3://$(S3_RELEASE_BUCKET_TESTING)/ro_volume.tar.gz
	rm -f /tmp/ro_volume.tar.gz
	cp site/public/robots.txt.disable site/public/robots.txt
	tar zcf /tmp/dist.tar.gz -C site dist
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/dist.tar.gz s3://$(S3_RELEASE_BUCKET_TESTING)/dist.tar.gz
	rm -f /tmp/dist.tar.gz site/public/robots.txt

.PHONY: upload-prod
upload-prod: update-challenges build build-static
	aws $(AWS_EXTRA_ARGS) s3 cp cmdchallenge/serve s3://$(S3_RELEASE_BUCKET_PROD)/serve
	tar zcf /tmp/ro_volume.tar.gz -C cmdchallenge ro_volume/
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/ro_volume.tar.gz s3://$(S3_RELEASE_BUCKET_PROD)/ro_volume.tar.gz
	rm -f /tmp/ro_volume.tar.gz
	tar zcf /tmp/dist.tar.gz -C site dist
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/dist.tar.gz s3://$(S3_RELEASE_BUCKET_PROD)/dist.tar.gz
	rm -f /tmp/dist.tar.gz

.PHONY: build
build:
	@echo "Building cmdchallenge ..."
	./bin/build-cmdchallenge

###################
# CMD Challenge
###################

.PHONY: test
test: push-image-cmd-testing
	cd $(DIR_CMDCHALLENGE); CMD_IMAGE_TAG=testing go test ./...

.PHONY: push-image-ci
push-image-ci: build-image-ci
	docker push $(CI_REGISTRY_IMAGE)/ci:latest

.PHONY: push-image-cmd-prod
push-image-cmd-prod: build-runcmd update-challenges tar-var
	bin/build-cmd-img prod
	rm -f var.tar.gz
.PHONY: push-image-cmd-testing
push-image-cmd-testing: build-runcmd update-challenges tar-var
	bin/build-cmd-img testing
	rm -f var.tar.gz

.PHONY: build-image-ci
build-image-ci:
	docker build -t $(CI_REGISTRY_IMAGE)/ci:latest -f Dockerfile-ci .

.PHONY: build-runcmd
build-runcmd:
	docker run --rm -v $(DIR_CMDRUNNER)/runcmd:/usr/src/app -w /usr/src/app nimlang/nim:1.4.0 nimble install -y
	docker run --rm -v $(DIR_CMDRUNNER)/oops:/usr/src/app -w /usr/src/app nimlang/nim:1.4.0 nimble install -y

.PHONY: build-test
build-test:
	docker run --rm -v $(DIR_CMDRUNNER)/test:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

.PHONY: update-challenges
update-challenges:
	./bin/update-challenges

.PHONY: build-static
build-static:
	cd $(DIR_SITE); npx vite build

.PHONY: tar-var
tar-var:
	cd $(DIR_CMDRUNNER); tar --exclude='.gitignore' --exclude='.gitkeep' -czf var.tar.gz var/

.PHONY: cmdshell
cmdshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDRUNNER)/ro_volume",target=/ro_volume  $(CI_REGISTRY_IMAGE)/cmd:latest bash

.PHONY: oopsshell
oopsshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDRUNNER)/ro_volume",target=/ro_volume $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest bash

.PHONY: clean
clean:
	rm -f $(DIR_CMDRUNNER)/ro_volume/ch/*
	rm -f $(PWD)/static/challenges/*
