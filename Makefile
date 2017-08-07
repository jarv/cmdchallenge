
REF=$(shell git rev-parse --short HEAD)

all: build-docker test

.PHONY: test
test: test-challenges test-runcmd

.PHONY: test-runcmd
test-runcmd:
	cd runcmd; go test

.PHONY: test-challenges
test-challenges:
	./bin/test_challenges

.PHONY: test-var
tar-var:
	tar -czf var.tar.gz var/

.PHONY: build-docker
build-docker: update-ro-volume gen-readme tar-var
	docker build -t registry.gitlab.com/jarv/cmdchallenge .

.PHONY: build-docker-stagin
build-docker-staging: update-ro-volume gen-readme tar-var
	docker build -t registry.gitlab.com/jarv/cmdchallenge:staging-$(REF) .

.PHONY: build-docker-prod
build-docker-prod: update-ro-volume gen-readme tar-var
	docker build -t registry.gitlab.com/jarv/cmdchallenge:prod-$(REF) .

clean:
	rm -f var.tar.gz

.PHONY: gen-readme
gen-readme:
	./bin/gen_readme

.PHONY: update-ro-volume
update-ro-volume:
	./bin/update-ro-volume

.PHONY: build-runcmd
build-runcmd:
	GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

.PHONY: build-runcmd-darwin
build-runcmd-darwin:
	go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go
