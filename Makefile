.PHONY: all test docker gen-readme update runcmd runcmd-darwin update-ro-volume test-runcmd test-challenges tar-var
REF=$(shell git rev-parse --short HEAD)

all: build-docker test

test: test-challenges test-runcmd

test-runcmd:
	cd runcmd; go test

test-challenges:
	./bin/test_challenges
tar-var:
	tar -czf var.tar.gz var/

build-docker: update-ro-volume gen-readme tar-var
	docker build -t registry.gitlab.com/jarv/cmdchallenge .

build-docker-staging: update-ro-volume gen-readme tar-var build-docker
	docker build -t registry.gitlab.com/jarv/cmdchallenge/staging .

build-docker-prod: update-ro-volume gen-readme tar-var build-docker
	docker build -t registry.gitlab.com/jarv/cmdchallenge/prod .

build-docker-ci:
	docker build -t registry.gitlab.com/jarv/cmdchallenge/ci-image -f Dockerfile-ci .
clean:
	rm -f var.tar.gz

gen-readme:
	./bin/gen_readme

update-ro-volume:
	./bin/update-ro-volume

build-runcmd:
	GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

build-runcmd-darwin:
	go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go
