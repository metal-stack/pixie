.PHONY: none pixiecore-docker

none:
	@echo "Use glide and the go tool for development"
	@echo "This makefile is just a shortcut for building docker containers."

pixiecore-docker:
	sudo docker build --rm -t danderson/pixiecore -f cmd/pixiecore/Dockerfile .
