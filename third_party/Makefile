.PHONY: ipxe

ipxe:
	rm -rf ipxe
	git clone git://git.ipxe.org/ipxe.git
	(cd ipxe && git rev-parse HEAD >COMMIT-ID)
	rm -rf ipxe/.git
	(cd ipxe/src && make bin/ipxe.pxe EMBED=../../../pixiecore/boot.ipxe)
	(cd ipxe/src && make bin-x86_64-efi/ipxe.efi EMBED=../../../pixiecore/boot.ipxe)
	(cd ipxe/src && make bin-i386-efi/ipxe.efi EMBED=../../../pixiecore/boot.ipxe)
	(cd ipxe && rm -rf bin && mkdir bin)
	mv -f ipxe/src/bin/ipxe.pxe ipxe/bin/ipxe.pxe
	mv -f ipxe/src/bin-x86_64-efi/ipxe.efi ipxe/bin/ipxe-x86_64.efi
	mv -f ipxe/src/bin-i386-efi/ipxe.efi ipxe/bin/ipxe-i386.efi
	go-bindata -o ipxe/ipxe-bin.go -pkg ipxe -nometadata -nomemcopy -prefix ipxe/bin/ ipxe/bin
	rm -rf ipxe/bin
	(cd ipxe/src && make clean)
