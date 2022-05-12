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
package cli // import "github.com/metal-stack/pixie/cli"

import (
	"fmt"
	"os"

	"github.com/metal-stack/pixie/pixiecore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Ipxe is the set of ipxe binaries for supported firmwares.
//
// Can be set externally before calling CLI(), and set/extended by
// commandline processing in CLI().
var Ipxe = map[pixiecore.Firmware][]byte{}

// CLI runs the Pixiecore commandline.
//
// This function always exits back to the OS when finished.
func CLI() {
	cobra.OnInitialize(initConfig)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	os.Exit(0)
}

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pixiecore",
	Short: "All-in-one network booting",
	Long:  `Pixiecore is a tool to make network booting easy.`,
}

func initConfig() {
	viper.SetEnvPrefix("pixiecore")
	viper.AutomaticEnv() // read in environment variables that match
}

func fatalf(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func serverConfigFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("debug", "d", false, "Log more things that aren't directly related to booting a recognized client")
	cmd.Flags().StringP("listen-addr", "l", "0.0.0.0", "IPv4 address to listen on")
	cmd.Flags().IntP("port", "p", 80, "Port to listen on for HTTP")
	cmd.Flags().String("metrics-listen-addr", "0.0.0.0", "IPv4 address of the metrics server to listen on")
	cmd.Flags().Int("metrics-port", 2113, "Metrics server port")
	cmd.Flags().Int("status-port", 0, "HTTP port for status information (can be the same as --port)")
	cmd.Flags().Bool("dhcp-no-bind", false, "Handle DHCP traffic without binding to the DHCP server port")
	cmd.Flags().String("ipxe-bios", "", "Path to an iPXE binary for BIOS/UNDI")
	cmd.Flags().String("ipxe-ipxe", "", "Path to an iPXE binary for chainloading from another iPXE")
	cmd.Flags().String("ipxe-efi32", "", "Path to an iPXE binary for 32-bit UEFI")
	cmd.Flags().String("ipxe-efi64", "", "Path to an iPXE binary for 64-bit UEFI")
}

func mustFile(path string) []byte {
	bs, err := os.ReadFile(path)
	if err != nil {
		fatalf("couldn't read file %q: %s", path, err)
	}

	return bs
}

func serverFromFlags(cmd *cobra.Command) *pixiecore.Server {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	addr, err := cmd.Flags().GetString("listen-addr")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	httpPort, err := cmd.Flags().GetInt("port")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	metricsAddr, err := cmd.Flags().GetString("metrics-listen-addr")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	metricsPort, err := cmd.Flags().GetInt("metrics-port")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	httpStatusPort, err := cmd.Flags().GetInt("status-port")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	dhcpNoBind, err := cmd.Flags().GetBool("dhcp-no-bind")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeBios, err := cmd.Flags().GetString("ipxe-bios")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeIpxe, err := cmd.Flags().GetString("ipxe-ipxe")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeEFI32, err := cmd.Flags().GetString("ipxe-efi32")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeEFI64, err := cmd.Flags().GetString("ipxe-efi64")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}

	if httpPort <= 0 {
		fatalf("HTTP port must be >0")
	}
	log, err := getLogger(debug)
	if err != nil {
		fatalf("Error creating logging: %s", err)
	}

	ret := &pixiecore.Server{
		Ipxe:           map[pixiecore.Firmware][]byte{},
		Log:            log,
		HTTPPort:       httpPort,
		HTTPStatusPort: httpStatusPort,
		MetricsPort:    metricsPort,
		MetricsAddress: metricsAddr,
		DHCPNoBind:     dhcpNoBind,
	}
	for fwtype, bs := range Ipxe {
		ret.Ipxe[fwtype] = bs
	}
	if ipxeBios != "" {
		ret.Ipxe[pixiecore.FirmwareX86PC] = mustFile(ipxeBios)
	}
	if ipxeIpxe != "" {
		ret.Ipxe[pixiecore.FirmwareX86Ipxe] = mustFile(ipxeIpxe)
	}
	if ipxeEFI32 != "" {
		ret.Ipxe[pixiecore.FirmwareEFI32] = mustFile(ipxeEFI32)
	}
	if ipxeEFI64 != "" {
		ret.Ipxe[pixiecore.FirmwareEFI64] = mustFile(ipxeEFI64)
		ret.Ipxe[pixiecore.FirmwareEFIBC] = ret.Ipxe[pixiecore.FirmwareEFI64]
	}
	if addr != "" {
		ret.Address = addr
	}

	return ret
}

func getLogger(debug bool) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zlog, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zlog.Sugar(), nil
}
