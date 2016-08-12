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

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.universe.tf/netboot/pixiecore"
)

// CLI runs the Pixiecore commandline.
//
// Takes a map of ipxe bootloader binaries for various architectures.
func CLI(ipxe map[pixiecore.Firmware][]byte) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var cfgFile string

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pixiecore",
	Short: "All-in-one network booting",
	Long:  `Pixiecore is a tool to make network booting easy.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading configuration file %q: %s\n", viper.ConfigFileUsed(), err)
			os.Exit(1)
		}
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.SetEnvPrefix("pixiecore")
	viper.AutomaticEnv() // read in environment variables that match
}

func fatalf(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func todo(msg string, args ...interface{}) {
	fatalf("TODO: "+msg, args...)
}
