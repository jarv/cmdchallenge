.PHONY: all test docker update runcmd runcmd-darwin update-challenges test-runcmd test-challenges tar-var push-image-cmd push-image-ci build-image-cmd build-image-ci
REF=$(shell git rev-parse --short HEAD)
PWD=$(shell pwd)
BASEDIR=$(CURDIR)
DIR_CMDCHALLENGE=$(CURDIR)/cmdchallenge
CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge
CI_COMMIT_TAG?=$(shell git rev-parse --short HEAD)
STATIC_OUTPUTDIR=$(BASEDIR)/static
AWS_PROFILE := cmdchallenge
DISTID_TESTING := E19XPJRE5YLRKA
S3_BUCKET_TESTING := testing.cmdchallenge.com

DISTID_PROD := E2UJHVXTJLOPCD
S3_BUCKET_PROD:= cmdchallenge.com

all: build-image-cmd test-challenges

##################
# Static site
##################

serve:
	cd static; python -m http.server 8000 --bind 127.0.0.1

serve_prod:
	./bin/simple-server prod

update:
	./bin/update-challenges-for-site
wsass:
	bundle exec sass --watch sass:static/css --style compressed

publish_testing: update-challenges
	cp static/robots.txt.disable static/robots.txt
	aws s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read --exclude "s/solutions/*" --delete
	aws --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	rm -f static/robots.txt

publish_testing_profile: update-challenges
	cp static/robots.txt.disable static/robots.txt
	aws --profile cmdchallenge s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read  --exclude "s/solutions/*"  --delete
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	rm -f static/robots.txt

publish_prod: update-challenges
	aws s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_PROD) --acl public-read --exclude "s/solutions/*" --delete
	aws --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_PROD) --paths '/*'

publish_prod_profile: update-challenges
	aws --profile cmdchallenge s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_PROD) --acl public-read  --exclude "s/solutions/*"  --delete
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_PROD) --paths '/*'

update-challenges:
	./bin/update-challenges

###################
# CMD Challenge
###################

test-runcmd:
	cd $(DIR_CMDCHALLENGE)/runcmd; go test

test-challenges:
	./bin/test_challenges

push-image-cmd: build-image-cmd
	docker push $(CI_REGISTRY_IMAGE)/cmd:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/cmd:latest
	docker push $(CI_REGISTRY_IMAGE)/cmd-no-bin:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest

push-image-ci: build-image-ci
	docker push $(CI_REGISTRY_IMAGE)/ci:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/ci:latest

build-image-cmd: build-runcmd update-challenges tar-var
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE)/cmd:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd:$(REF) .
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE)/cmd-no-bin:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd-no-bin:$(REF) -f Dockerfile-no-bin .
	rm -f var.tar.gz

build-image-ci:
	docker build -t $(CI_REGISTRY_IMAGE)/ci:latest \
		--tag $(CI_REGISTRY_IMAGE)/ci:$(CI_COMMIT_TAG) -f Dockerfile-ci .

build-runcmd:
	docker run --rm -v $(DIR_CMDCHALLENGE)/runcmd:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y
	docker run --rm -v $(DIR_CMDCHALLENGE)/oops:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

build-test:
	docker run --rm -v $(DIR_CMDCHALLENGE)/test:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

update-challenges:
	./bin/update-challenges

tar-var:
	cd $(DIR_CMDCHALLENGE); tar --exclude='.gitignore' --exclude='.gitkeep' -czf var.tar.gz var/

cmdshell:
	docker run -it --privileged --mount type=bind,source="$(PWD)/cmdchallenge/ro_volume",target=/ro_volume  registry.gitlab.com/jarv/cmdchallenge/cmd:latest bash

oopsshell:
	docker run -it --privileged --mount type=bind,source="$(PWD)/cmdchallenge/ro_volume",target=/ro_volume  registry.gitlab.com/jarv/cmdchallenge/cmd-no-bin:latest bash

testshell:
	docker run -it --privileged --mount type=bind,source="$(PWD)/cmdchallenge/test",target=/tmp registry.gitlab.com/jarv/cmdchallenge/cmd:latest bash

clean:
	rm -f $(PWD)/cmdchallenge/ro_volume/ch/*
	rm -f $(PWD)/static/challenges/*
