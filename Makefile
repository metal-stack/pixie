.PHONY: none pixiecore pixiecore-git

none:
	@echo "Use glide and the go tool for development"
	@echo "This makefile is just a shortcut for building docker containers."

pixiecore:
	sudo docker build --rm -t danderson/pixiecore -f dockerfiles/pixiecore/Dockerfile .

pixiecore-git:
	sudo docker build --rm -t danderson/pixiecore dockerfiles/pixiecore
