test: docker
	./bin/test_challenges
docker: gen_readme
	tar -hczf var.tar.gz var/
	docker build -t cmdline .
	rm -f var.tar.gz

gen_readme:
	./bin/gen_readme

all: test
