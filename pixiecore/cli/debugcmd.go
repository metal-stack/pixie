// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

var (
	debugCmd = &cobra.Command{
		Use:    "debug",
		Short:  "Internal debugging commands",
		Hidden: true,
	}
	dumpIpxeCmd = &cobra.Command{
		Use:   "dump-ipxe",
		Short: "Dump the builtin ipxe binaries to disk",
		Run: func(cmd *cobra.Command, args []string) {
			for fwtype, bs := range Ipxe {
				path := fmt.Sprintf("builtin-ipxe-%d", fwtype)
				if err := ioutil.WriteFile(path, bs, 0644); err != nil {
					fmt.Printf("Error writing %s: %s\n", path, err)
				} else {
					fmt.Println("Wrote", path)
				}
			}
		},
	}
)

func init() {
	debugCmd.AddCommand(dumpIpxeCmd)
	rootCmd.AddCommand(debugCmd)
}
