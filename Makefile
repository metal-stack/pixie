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

all: ipxe pixie

.PHONY: pixie
pixie: test
	go build -tags netgo,osusergo \
		 -ldflags "$(LINKMODE) -X 'github.com/metal-stack/v.Version=$(VERSION)' \
								   -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
								   -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
								   -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'" \
	   -o build/pixie github.com/metal-stack/pixie/cmd
	strip build/pixie

.PHONY: test
test:
	$(GOMODULECMD) test ./...
	$(GOMODULECMD) test -race ./...

.PHONY: lint
lint:
	$(GOMODULECMD) vet ./...


IPXE_COMMIT_SHA := $(shell cat ipxe/IPXE_COMMIT_SHA)

# Note: requires liblzma-dev package installed
.PHONY: ipxe
ipxe:
	rm -rf ipxe/ipxe
	(cd ipxe \
		&& git clone https://github.com/ipxe/ipxe.git \
		&& cd ipxe \
		&& git checkout $(IPXE_COMMIT_SHA) \
	)
	cp ipxe/branding.h ipxe/ipxe/src/config/local/branding.h
	(cd ipxe/ipxe/src &&\
		make bin/ipxe.pxe bin/undionly.kpxe bin-x86_64-efi/ipxe.efi bin-x86_64-efi/snponly.efi bin-i386-efi/ipxe.efi EMBED=../../../pixiecore/boot.ipxe)
	(rm -rf ipxe/ipxe/bin && mkdir ipxe/ipxe/bin)
	mv -f ipxe/ipxe/src/bin/ipxe.pxe ipxe/ipxe/bin/ipxe.pxe
	mv -f ipxe/ipxe/src/bin-x86_64-efi/snponly.efi ipxe/ipxe/bin/snponly.efi
	mv -f ipxe/ipxe/src/bin/undionly.kpxe ipxe/ipxe/bin/undionly.kpxe
	mv -f ipxe/ipxe/src/bin-x86_64-efi/ipxe.efi ipxe/ipxe/bin/ipxe-x86_64.efi
	mv -f ipxe/ipxe/src/bin-i386-efi/ipxe.efi ipxe/ipxe/bin/ipxe-i386.efi
	(cd ipxe/ipxe/src && make veryclean)
