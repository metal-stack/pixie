GOCMD:=go

# Local customizations to the above.
ifneq ($(wildcard Makefile.defaults),)
include Makefile.defaults
endif

all:
	$(error Please request a specific thing, there is no default target)

.PHONY: ci-config
ci-config:
	(cd .circleci && go run gen-config.go >config.yml)

.PHONY: ci-prepare
ci-prepare:
	$(GOCMD) get -u github.com/golang/dep/cmd/dep
	$(GOCMD) get -u github.com/estesp/manifest-tool
	dep ensure

.PHONY: ci-build
ci-build:
	$(GOCMD) install -v ./cmd/pixiecore

.PHONY: ci-test
ci-test:
	$(GOCMD) test ./...
	$(GOCMD) test -race ./...

.PHONY: ci-lint
ci-lint:
	$(GOCMD) get -u github.com/alecthomas/gometalinter
	gometalinter --install golint
	gometalinter --deadline=1m --disable-all --enable=gofmt --enable=golint --enable=vet --enable=vetshadow --enable=structcheck --enable=unconvert --vendor ./...

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
