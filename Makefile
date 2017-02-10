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
all: test
