// Copyright © 2016 David Anderson <dave@natulte.net>
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

import "github.com/spf13/cobra"

var quickCmd = &cobra.Command{
	Use:   "quick recipe [settings...]",
	Short: "Boot an OS from a list",
	Long: `This ends up working the same as the simple boot command, but saves
you having to find the kernels and ramdisks for popular OSes.

TODO: better help here
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fatalf("you must specify at least a recipe")
		}
		recipe := args[0]
		todo("run in quick mode with recipe=%s", recipe)
	},
}

func init() {
	//rootCmd.AddCommand(quickCmd)

	// TODO: some kind of caching support where quick OSes get
	// downloaded locally, so you don't have to fetch from a remote
	// server on every boot attempt.
}
