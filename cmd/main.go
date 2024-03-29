// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/metal-stack/pixie/ipxe"
	"github.com/metal-stack/pixie/pixiecore"
	"github.com/metal-stack/pixie/pixiecore/cli"
)

func main() {
	cli.Ipxe[pixiecore.FirmwareX86PC] = ipxe.MustGet("undionly.kpxe")
	cli.Ipxe[pixiecore.FirmwareEFI32] = ipxe.MustGet("ipxe-i386.efi")
	cli.Ipxe[pixiecore.FirmwareEFI64] = ipxe.MustGet("ipxe-x86_64.efi")
	cli.Ipxe[pixiecore.FirmwareEFIBC] = ipxe.MustGet("ipxe-x86_64.efi")
	cli.Ipxe[pixiecore.FirmwareX86Ipxe] = ipxe.MustGet("ipxe.pxe")
	cli.CLI()
}
