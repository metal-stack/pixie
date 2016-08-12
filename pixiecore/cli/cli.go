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

// Package cli implements the commandline interface for Pixiecore.
package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.universe.tf/netboot/pixiecore"
)

// Ipxe is the set of ipxe binaries for supported firmwares.
//
// Can be set externally before calling CLI(), and set/extended by
// commandline processing in CLI().
var Ipxe = map[pixiecore.Firmware][]byte{}

// CLI runs the Pixiecore commandline.
//
// Takes a map of ipxe bootloader binaries for various architectures.
func CLI() {
	// The ipxe firmware flags need to be set outside init(), so that
	// the default flag value is computed appropriately based on
	// whether the caller preseeded Ipxe.
	rootCmd.PersistentFlags().Var(ipxeFirmwareFlag(pixiecore.FirmwareX86PC), "ipxe-bios", "iPXE binary for BIOS/UNDI")
	rootCmd.PersistentFlags().Var(ipxeFirmwareFlag(pixiecore.FirmwareEFI32), "ipxe-efi32", "iPXE binary for 32-bit UEFI")
	rootCmd.PersistentFlags().Var(ipxeFirmwareFlag(pixiecore.FirmwareEFI64), "ipxe-efi64", "iPXE binary for 64-bit UEFI")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pixiecore",
	Short: "All-in-one network booting",
	Long:  `Pixiecore is a tool to make network booting easy.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

type ipxeFirmwareFlag pixiecore.Firmware

func (iff ipxeFirmwareFlag) String() string {
	if Ipxe[pixiecore.Firmware(iff)] != nil {
		return "<builtin>"
	}
	return ""
}

func (iff ipxeFirmwareFlag) Set(v string) error {
	bs, err := ioutil.ReadFile(v)
	if err != nil {
		return fmt.Errorf("couldn't read ipxe binary %q: %s", v, err)
	}

	Ipxe[pixiecore.Firmware(iff)] = bs

	return nil
}

func (ipxeFirmwareFlag) Type() string {
	return "filename"
}

var cfgFile string

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
