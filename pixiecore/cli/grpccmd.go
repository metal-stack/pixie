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

import (
	"fmt"
	"os"

	"github.com/metal-stack/pixie/pixiecore"
	"github.com/spf13/cobra"
)

var grpcCmd = &cobra.Command{
	Use:   "grpc server",
	Short: "Boot machines using instructions from a grpc server",
	Long: `API mode is a "PXE to GRPC" translator. Whenever Pixiecore sees a
machine trying to PXE boot, it will ask a remote grpc server
what to do. The API server can tell Pixiecore to ignore the machine,
or tell it what to boot.

It is your responsibility to implement or run a server that implements
the Pixiecore boot API. The specification can be found at <TODO>.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := serverFromFlags(cmd)

		metalAPIConfig, err := getMetalAPIConfig(cmd)
		if err != nil {
			fatalf("unable to create metal-api config: %s", err)
		}
		client, err := pixiecore.NewGrpcClient(s.Log, metalAPIConfig)
		if err != nil {
			fatalf("unable to create grpc client: %s", err)
		}
		partition, err := cmd.Flags().GetString("grpc-cert")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		booter, err := pixiecore.GRPCBooter(s.Log, client, partition, metalAPIConfig)
		if err != nil {
			fatalf("unable to create grpc booter: %s", err)
		}
		s.Booter = booter
		s.MetalConfig = metalAPIConfig

		fmt.Println(s.Serve())
	}}

func init() {
	rootCmd.AddCommand(grpcCmd)
	serverConfigFlags(grpcCmd)

	grpcCmd.Flags().String("partitionID", "", "id of the partition this instance of pixie is running")

	grpcCmd.Flags().String("pixie-api-url", "", "base url of pixie itself")

	grpcCmd.Flags().String("grpc-ca-cert", "", "Path to the grpc ca cert file")
	grpcCmd.Flags().String("grpc-cert", "", "Path to the grpc client cert file")
	grpcCmd.Flags().String("grpc-key", "", "Path to the grpc client key file")
	grpcCmd.Flags().String("grpc-address", "", "address of the grpc server")
	grpcCmd.Flags().String("metal-api-view-hmac", "", "hmac with metal-api view access")
	grpcCmd.Flags().String("metal-api-url", "", "url to access metal-api")
	grpcCmd.Flags().Bool("metal-hammer-debug", true, "set metal-hammer to debug")

	// must(grpcCmd.MarkFlagRequired("pixie-url"))
	// must(grpcCmd.MarkFlagRequired("partitionID"))
	// must(grpcCmd.MarkFlagRequired("grpc-ca-cert"))
	// must(grpcCmd.MarkFlagRequired("grpc-cert"))
	// must(grpcCmd.MarkFlagRequired("grpc-key"))
	// must(grpcCmd.MarkFlagRequired("metal-api-view-hmac"))
	// must(grpcCmd.MarkFlagRequired("metal-api-url"))

}

// func must(err error) {
// 	if err != nil {
// 		panic(err.Error())
// 	}
// }
func getMetalAPIConfig(cmd *cobra.Command) (*pixiecore.MetalConfig, error) {
	grpcCACertFile, err := cmd.Flags().GetString("grpc-ca-cert")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	caCert, err := os.ReadFile(grpcCACertFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read ca-cert %w", err)
	}

	grpcClientCertFile, err := cmd.Flags().GetString("grpc-cert")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	clientCert, err := os.ReadFile(grpcClientCertFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read cert %w", err)
	}

	grpcClientKeyFile, err := cmd.Flags().GetString("grpc-key")
	if err != nil {
		return nil, fmt.Errorf("unable to read key %w", err)
	}
	clientKey, err := os.ReadFile(grpcClientKeyFile)
	if err != nil {
		return nil, err
	}
	grpcAddress, err := cmd.Flags().GetString("grpc-address")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}

	hmac, err := cmd.Flags().GetString("metal-api-view-hmac")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	metalAPIUrl, err := cmd.Flags().GetString("metal-api-url")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	pixieAPIUrl, err := cmd.Flags().GetString("pixie-api-url")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	metalHammerDebug, err := cmd.Flags().GetBool("metal-hammer-debug")
	if err != nil {
		return nil, fmt.Errorf("Error reading flag: %w", err)
	}
	return &pixiecore.MetalConfig{
		Debug:       metalHammerDebug,
		GRPCAddress: grpcAddress,
		MetalAPIUrl: metalAPIUrl,
		PixieAPIURL: pixieAPIUrl,
		CACert:      string(caCert),
		Cert:        string(clientCert),
		Key:         string(clientKey),
		HMAC:        hmac,
	}, nil
}
