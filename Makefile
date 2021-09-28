CI_COMMIT_SHORT_SHA?=$(shell git rev-parse --short HEAD)
CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge
PWD=$(shell pwd)
DATE_TS=$(shell date -u +%Y%m%d%H%M%S)
AWS_EXTRA_ARGS=$(shell [ -z $$CI ] && echo "--profile cmdchallenge" || echo "")
BASEDIR=$(CURDIR)
DIR_CMDRUNNER=$(CURDIR)/cmdrunner
DIR_CMDCHALLENGE=$(CURDIR)/cmdchallenge
STATIC_OUTPUTDIR=$(BASEDIR)/static
AWS_PROFILE := cmdchallenge
DISTID_TESTING := E19XPJRE5YLRKA
S3_BUCKET_TESTING := testing.cmdchallenge.com
S3_BUCKET_PROD:= cmdchallenge.com

S3_RELEASE_BUCKET_PROD := prod-cmd-release
S3_RELEASE_BUCKET_TESTING := testing-cmd-release

DISTID_PROD := E2UJHVXTJLOPCD

all: upload-testing

.PHONY: upload-testing
upload-testing: build update-challenges
	aws $(AWS_EXTRA_ARGS) s3 cp cmdchallenge/serve s3://$(S3_RELEASE_BUCKET_TESTING)/serve
	tar zcf /tmp/ro_volume.tar.gz -C cmdchallenge ro_volume/
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/ro_volume.tar.gz s3://$(S3_RELEASE_BUCKET_TESTING)/ro_volume.tar.gz
	rm -f /tmp/ro_volume.tar.gz

.PHONY: upload-prod
upload-prod: build update-challenges
	aws $(AWS_EXTRA_ARGS) s3 cp cmdchallenge/serve s3://$(S3_RELEASE_BUCKET_PROD)/serve.$(CI_COMMIT_SHORT_SHA)
	tar zcf /tmp/ro_volume.tar.gz -C cmdchallenge ro_volume/
	aws $(AWS_EXTRA_ARGS) s3 cp /tmp/ro_volume.tar.gz s3://$(S3_RELEASE_BUCKET_PROD)/ro_volume.tar.gz.$(CI_COMMIT_SHORT_SHA)
	rm -f /tmp/ro_volume.tar.gz

.PHONY: build
build:
	@echo "Building cmdchallenge ..."
	./bin/build-cmdchallenge

.PHONY: serve
serve:
	cd static; python -m http.server 8000 --bind 127.0.0.1

.PHONY: server_prod
serve_prod:
	./bin/simple-server prod

.PHONY: wsass
wsass:
	bundle exec sass --watch sass:static/css --style compressed

.PHONY: publish-testing
publish-testing: update-challenges cache-bust-index
	cp static/robots.txt.disable static/robots.txt
	aws $(AWS_EXTRA_ARGS) s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read --delete --cache-control max-age=604800
	aws $(AWS_EXTRA_ARGS) s3 cp s3://$(S3_BUCKET_TESTING)/index.html s3://$(S3_BUCKET_TESTING)/index.html --metadata-directive REPLACE --cache-control max-age=0,no-cache,no-store,must-revalidate --content-type text/html --acl public-read
	aws $(AWS_EXTRA_ARGS) s3 cp s3://$(S3_BUCKET_TESTING)/challenges/challenges.json s3://$(S3_BUCKET_TESTING)/challenges/challenges.json --metadata-directive REPLACE --cache-control max-age=0,no-cache,no-store,must-revalidate --content-type application/json --acl public-read
	aws $(AWS_EXTRA_ARGS) --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	rm -f static/robots.txt
	git checkout static/index.html

.PHONY: publish-prod
publish-prod: update-challenges cache-bust-index
	aws $(AWS_EXTRA_ARGS) s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_PROD) --acl public-read --delete --cache-control max-age=604800
	aws $(AWS_EXTRA_ARGS) s3 cp s3://$(S3_BUCKET_PROD)/index.html s3://$(S3_BUCKET_PROD)/index.html --metadata-directive REPLACE --cache-control max-age=0,no-cache,no-store,must-revalidate --content-type text/html --acl public-read
	aws $(AWS_EXTRA_ARGS) s3 cp s3://$(S3_BUCKET_PROD)/challenges/challenges.json s3://$(S3_BUCKET_PROD)/challenges/challenges.json --metadata-directive REPLACE --cache-control max-age=0,no-cache,no-store,must-revalidate --content-type application/json --acl public-read
	aws $(AWS_EXTRA_ARGS) --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_PROD) --paths '/*'
	git checkout static/index.html


###################
# CMD Challenge
###################

.PHONY: test
test: push-image-cmd
	cd $(DIR_CMDCHALLENGE); CMD_IMAGE_TAG=$(CI_COMMIT_SHORT_SHA) go test ./...

.PHONY: push-image-cmd
push-image-cmd: build-image-cmd
	docker push $(CI_REGISTRY_IMAGE)/cmd:$(CI_COMMIT_SHORT_SHA)
	docker push $(CI_REGISTRY_IMAGE)/cmd:latest
	docker push $(CI_REGISTRY_IMAGE)/cmd-no-bin:$(CI_COMMIT_SHORT_SHA)
	docker push $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest

.PHONY: push-image-ci
push-image-ci: build-image-ci
	docker push $(CI_REGISTRY_IMAGE)/ci:$(CI_COMMIT_SHORT_SHA)
	docker push $(CI_REGISTRY_IMAGE)/ci:latest

.PHONY: build-image-cmd
build-image-cmd: build-runcmd update-challenges tar-var
	cd $(DIR_CMDRUNNER); docker build -t $(CI_REGISTRY_IMAGE)/cmd:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd:$(CI_COMMIT_SHORT_SHA) .
	cd $(DIR_CMDRUNNER); docker build -t $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd-no-bin:$(CI_COMMIT_SHORT_SHA) -f Dockerfile-no-bin .
	rm -f var.tar.gz

.PHONY: build-image-ci
build-image-ci:
	docker build -t $(CI_REGISTRY_IMAGE)/ci:latest \
		--tag $(CI_REGISTRY_IMAGE)/ci:$(CI_COMMIT_SHORT_SHA) -f Dockerfile-ci .

.PHONY: build-runcmd
build-runcmd:
	docker run --rm -v $(DIR_CMDRUNNER)/runcmd:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y
	docker run --rm -v $(DIR_CMDRUNNER)/oops:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

.PHONY: build-test
build-test:
	docker run --rm -v $(DIR_CMDRUNNER)/test:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

.PHONY: update-challenges
update-challenges:
	./bin/update-challenges

.PHONY: tar-var
tar-var:
	cd $(DIR_CMDRUNNER); tar --exclude='.gitignore' --exclude='.gitkeep' -czf var.tar.gz var/

.PHONY: cmdcshell
cmdshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDRUNNER)/ro_volume",target=/ro_volume  $(CI_REGISTRY_IMAGE)/cmd:latest bash

.PHONY: oopsshell
oopsshell:
	docker run -it --privileged --mount type=bind,source="$(DIR_CMDRUNNER)/ro_volume",target=/ro_volume $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest bash

.PHONY: clean
clean:
	rm -f $(DIR_CMDRUNNER)/ro_volume/ch/*
	rm -f $(PWD)/static/challenges/*

.PHONY: cache-bust-index
cache-bust-index:
	sed -i -e "s/\.css\"/.css?$(DATE_TS)\"/" static/index.html
	sed -i -e "s/\.js\"/.js?$(DATE_TS)\"/" static/index.html
