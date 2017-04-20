.PHONY: all test docker gen_readme update runcmd

test: docker
	./bin/test_challenges
docker: gen_readme
	tar -czf var.tar.gz var/
	docker build -t cmdline .
	rm -f var.tar.gz
	docker save cmdline > img.tar

gen_readme:
	./bin/gen_readme
update:
	./bin/update
runcmd:
	go build -o ./ro_volume/runcmd-darwin ./runcmd/runcmd.go ./runcmd/challenges.go
	GOOS=linux GOARCH=amd64 go build -o ./ro_volume/runcmd ./runcmd/runcmd.go ./runcmd/challenges.go

all: test
