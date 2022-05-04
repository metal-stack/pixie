SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date -Iseconds)
VERSION := $(or ${VERSION},devel)

GOCMD:=go
GOMODULECMD:=GO111MODULE=on go
LINKMODE := -extldflags '-static -s -w'

# Local customizations to the above.
ifneq ($(wildcard Makefile.defaults),)
include Makefile.defaults
endif

all: pixiecore

.PHONY: pixiecore
pixiecore:
	go build -tags netgo,osusergo,urfave_cli_no_docs \
		 -ldflags "$(LINKMODE) -X 'github.com/metal-stack/v.Version=$(VERSION)' \
								   -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
								   -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
								   -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'" \
	   -o bin/pixiecore github.com/metal-stack/pixiecore/cmd/pixiecore
	strip bin/pixiecore

.PHONY: ci-prepare
ci-prepare:
	$(GOCMD) get -u github.com/estesp/manifest-tool

.PHONY: build
build:
	$(GOMODULECMD) install -v ./cmd/pixiecore

.PHONY: test
test:
	$(GOMODULECMD) test ./...
	$(GOMODULECMD) test -race ./...

.PHONY: lint
lint:
	$(GOMODULECMD) tool vet .

REGISTRY=pixiecore
TAG=dev
.PHONY: ci-push-images
ci-push-images:
	make -f Makefile.inc push GOARCH=amd64   TAG=$(TAG)-amd64   BINARY=pixiecore REGISTRY=$(REGISTRY)
	make -f Makefile.inc push GOARCH=arm     TAG=$(TAG)-arm     BINARY=pixiecore REGISTRY=$(REGISTRY)
	make -f Makefile.inc push GOARCH=arm64   TAG=$(TAG)-arm64   BINARY=pixiecore REGISTRY=$(REGISTRY)
	make -f Makefile.inc push GOARCH=ppc64le TAG=$(TAG)-ppc64le BINARY=pixiecore REGISTRY=$(REGISTRY)
	make -f Makefile.inc push GOARCH=s390x   TAG=$(TAG)-s390x   BINARY=pixiecore REGISTRY=$(REGISTRY)
	manifest-tool push from-args --platforms linux/amd64,linux/arm,linux/arm64,linux/ppc64le,linux/s390x --template $(REGISTRY)/pixiecore:$(TAG)-ARCH --target $(REGISTRY)/pixiecore:$(TAG)

.PHONY: ci-config
ci-config:
	(cd .circleci && go run gen-config.go >config.yml)

.PHONY: update-ipxe
update-ipxe:
	# rm -rf third_party/ipxe
	# (cd third_party && git clone git://git.ipxe.org/ipxe.git)
	# (cd third_party/ipxe && git rev-parse HEAD >COMMIT-ID)
	# rm -rf third_party/ipxe/.git
	(cd third_party/ipxe/src &&\
		make bin/ipxe.pxe bin/undionly.kpxe bin-x86_64-efi/ipxe.efi bin-i386-efi/ipxe.efi EMBED=../../../pixiecore/boot.ipxe)
	(rm -rf third_party/ipxe/bin && mkdir third_party/ipxe/bin)
	mv -f third_party/ipxe/src/bin/ipxe.pxe third_party/ipxe/bin/ipxe.pxe
	mv -f third_party/ipxe/src/bin/undionly.kpxe third_party/ipxe/bin/undionly.kpxe
	mv -f third_party/ipxe/src/bin-x86_64-efi/ipxe.efi third_party/ipxe/bin/ipxe-x86_64.efi
	mv -f third_party/ipxe/src/bin-i386-efi/ipxe.efi third_party/ipxe/bin/ipxe-i386.efi
	go-bindata -o third_party/ipxe/ipxe-bin.go -pkg ipxe -nometadata -nomemcopy -prefix third_party/ipxe/bin/ third_party/ipxe/bin
	gofmt -s -w third_party/ipxe/ipxe-bin.go
	rm -rf third_party/ipxe/bin
	(cd third_party/ipxe/src && make veryclean)
