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

var apiCmd = &cobra.Command{
	Use:   "api server",
	Short: "Boot machines using instructions from one or more API servers",
	Long: `API mode is a "PXE to HTTP" translator. Whenever Pixiecore sees a
machine trying to PXE boot, it will ask a remote HTTP(S) API server
what to do. The API server can tell Pixiecore to ignore the machine,
or tell it what to boot.

It is your responsibility to implement or run a server that implements
the Pixiecore boot API. The specification can be found at <TODO>.`,
	Run: func(cmd *cobra.Command, args []string) { todo("api called") }}

func init() {
	rootCmd.AddCommand(apiCmd)
	// TODO: SSL cert flags for both client and server auth.
}
