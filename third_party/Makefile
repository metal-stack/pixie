.PHONY: ipxe

ipxe:
	rm -rf ipxe
	git clone git://git.ipxe.org/ipxe.git
	(cd ipxe && git rev-parse HEAD >COMMIT-ID)
	rm -rf ipxe/.git
	(cd ipxe/src && make bin/undionly.kpxe bin-x86_64-efi/ipxe.efi bin-i386-efi/ipxe.efi)
	(cd ipxe && rm -rf bin && mkdir bin)
	mv -f ipxe/src/bin/undionly.kpxe ipxe/bin/undionly.kpxe
	mv -f ipxe/src/bin-x86_64-efi/ipxe.efi ipxe/bin/ipxe-x86_64.efi
	mv -f ipxe/src/bin-i386-efi/ipxe.efi ipxe/bin/ipxe-i386.efi
	go-bindata -o ipxe/ipxe-bin.go -pkg ipxe -nometadata -nomemcopy -prefix ipxe/bin/ ipxe/bin
	rm -rf ipxe/bin
	(cd ipxe/src && make clean)
