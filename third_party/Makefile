# Not requires liblzma-dev package installed

.PHONY: ipxe
ipxe:
	rm -rf ipxe
	git clone https://github.com/ipxe/ipxe.git
	(cd ipxe && git rev-parse HEAD >COMMIT-ID)
	rm -rf ipxe/.git
	(cd ipxe/src &&\
		make bin/ipxe.pxe bin/undionly.kpxe bin-x86_64-efi/ipxe.efi bin-i386-efi/ipxe.efi EMBED=../../../pixiecore/boot.ipxe)
	(cd ipxe && rm -rf bin && mkdir bin)
	mv -f ipxe/src/bin/ipxe.pxe ipxe/bin/ipxe.pxe
	mv -f ipxe/src/bin/undionly.kpxe ipxe/bin/undionly.kpxe
	mv -f ipxe/src/bin-x86_64-efi/ipxe.efi ipxe/bin/ipxe-x86_64.efi
	mv -f ipxe/src/bin-i386-efi/ipxe.efi ipxe/bin/ipxe-i386.efi
	(cd ipxe/src && make clean)
