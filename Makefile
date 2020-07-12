.PHONY: all test docker update runcmd runcmd-darwin update-challenges test-runcmd test-challenges tar-var push-image-cmd push-image-ci build-image-cmd build-image-ci
REF=$(shell git rev-parse --short HEAD)
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

all: build-image-cmd-testing test-challenges

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
	rm -f static/robots.txt

publish_testing_profile: gen-deps
	cp static/robots.txt.disable static/robots.txt
	aws --profile cmdchallenge s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_TESTING) --acl public-read  --exclude "s/solutions/*"  --delete
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_TESTING) --paths '/*'
	rm -f static/robots.txt

publish_prod: gen-deps
	aws s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_PROD) --acl public-read --exclude "s/solutions/*" --delete
	aws --region us-east-1 cloudfront create-invalidation --distribution-id $(DISTID_PROD) --paths '/*'

publish_prod_profile: gen-deps
	aws --profile cmdchallenge s3 sync $(STATIC_OUTPUTDIR)/ s3://$(S3_BUCKET_PROD) --acl public-read  --exclude "s/solutions/*"  --delete
	aws --region us-east-1 --profile $(AWS_PROFILE) cloudfront create-invalidation --distribution-id $(DISTID_PROD) --paths '/*'

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

push-image-cmd-testing: build-image-cmd-testing
	docker push $(CI_REGISTRY_IMAGE)/cmd:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/cmd:testing

push-image-ci: build-image-ci
	docker push $(CI_REGISTRY_IMAGE)/ci:$(REF)
	docker push $(CI_REGISTRY_IMAGE)/ci:latest

build-image-cmd: build-runcmd update-challenges tar-var
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE)/cmd:latest \
		--tag $(CI_REGISTRY_IMAGE)/cmd:$(REF) .
	rm -f var.tar.gz

build-image-cmd-testing: build-runcmd update-challenges tar-var
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE)/cmd:testing \
		--tag $(CI_REGISTRY_IMAGE)/cmd:$(REF) .
	rm -f var.tar.gz

build-image-ci:
	docker build -t $(CI_REGISTRY_IMAGE)/ci:latest \
		--tag $(CI_REGISTRY_IMAGE)/ci:$(CI_COMMIT_TAG) -f Dockerfile-ci .

build-runcmd-golang:
	cd $(DIR_CMDCHALLENGE); GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

build-runcmd-golang-darwin:
	cd $(DIR_CMDCHALLENGE); go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go

build-runcmd:
	docker run --rm -v $(DIR_CMDCHALLENGE)/runcmd:/usr/src/app -w /usr/src/app nimlang/nim nimble install -y

update-challenges:
	./bin/update-challenges

tar-var:
	cd $(DIR_CMDCHALLENGE); tar --exclude='.gitignore' -czf var.tar.gz var/
