.PHONY: all test docker gen-readme update runcmd runcmd-darwin update-ro-volume test-runcmd test-challenges tar-var
REF=$(shell git rev-parse --short HEAD)
DIR_CMDCHALLENGE=cmdchallenge
CI_REGISTRY_IMAGE?=registry.gitlab.com/jarv/cmdchallenge

all: test-runcmd build-docker test-challenges


###################

test-runcmd:
	cd $(DIR_CMDCHALLENGE)/runcmd; go test

test-challenges:
	./bin/test_challenges

test-gitlab-dnd:
	./bin/test_gitlab_dnd.py

build-docker: build-runcmd update-ro-volume gen-readme tar-var
	cd $(DIR_CMDCHALLENGE); docker build -t $(CI_REGISTRY_IMAGE):latest \
		--tag $(CI_REGISTRY_IMAGE):$(REF) \
	    --tag $(CI_REGISTRY_IMAGE):latest .; rm -f var.tar.gz

build-runcmd:
	cd $(DIR_CMDCHALLENGE); GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

build-runcmd-darwin:
	cd $(DIR_CMDCHALLENGE); go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go

update-ro-volume:
	./bin/update-ro-volume

gen-readme:
	./bin/gen_readme

tar-var:
	cd $(DIR_CMDCHALLENGE); tar -czf var.tar.gz var/
