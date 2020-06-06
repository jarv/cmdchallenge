.PHONY: all test docker update runcmd runcmd-darwin update-challenges test-runcmd test-challenges tar-var push-image-cmd push-image-ci build-image-cmd build-image-ci
REF=$(shell git rev-parse --short HEAD)
BASEDIR=$(CURDIR)
DIR_CMDCHALLENGE=$(CURDIR)/cmdchallenge
CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge
CI_COMMIT_TAG?=$(shell git rev-parse --short HEAD)
STATIC_OUTPUTDIR=$(BASEDIR)/static
AWS_PROFILE := cmdchallenge
DISTID_TESTING := E19XPJRE5YLRKA
DISTID_TESTING_API := E3T11IZ0ZVPJVT
S3_BUCKET_TESTING := testing.cmdchallenge.com
S3_BUCKET := cmdchallenge.com

all: test-runcmd build-image-cmd test-challenges

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

publish_testing: gen-deps
	cp static/robots.txt.disable static/robots.txt
	aws s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read --exclude "s/solutions/*" --delete
	aws --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	aws --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_TESTING_API) --paths '/*'
	rm -f static/robots.txt

publish_testing_profile: gen-deps
	cp static/robots.txt.disable static/robots.txt
	aws --profile cmdchallenge s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read  --exclude "s/solutions/*"  --delete
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_TESTING_API) --paths '/*'
	rm -f static/robots.txt

gen-deps:
	./bin/gen-deps

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

push-image-ci: build-image-ci
	docker push $(CI_REGISTRY_IMAGE)/ci:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/ci:latest

build-image-cmd: build-runcmd update-challenges tar-var
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE)/cmd:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd:$(REF) .
	rm -f var.tar.gz

build-image-ci:
	docker build -t $(CI_REGISTRY_IMAGE)/ci:latest \
		--tag $(CI_REGISTRY_IMAGE)/ci:$(CI_COMMIT_TAG) -f Dockerfile-ci .

build-runcmd:
	cd $(DIR_CMDCHALLENGE); GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

build-runcmd-darwin:
	cd $(DIR_CMDCHALLENGE); go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go

update-challenges:
	./bin/update-challenges

tar-var:
	cd $(DIR_CMDCHALLENGE); tar -czf var.tar.gz var/

shellcheck:
	find cmdchallenge/ro_volume/cmdtests/ cmdchallenge/ro_volume/randomizers/ -type f | xargs -r shellcheck -e SC1090,SC1091

shfmt:
	find cmdchallenge/ro_volume/cmdtests/ cmdchallenge/ro_volume/randomizers/ -type f | xargs -r shellcheck -e SC1090,SC1091
