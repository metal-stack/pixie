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
	"net/url"
	"os"
	"strings"

	"github.com/metal-stack/pixie/api"
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
		partition, err := cmd.Flags().GetString("partition")
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

	grpcCmd.Flags().String("partition", "", "id of the partition this instance of pixie is running")

	grpcCmd.Flags().String("pixie-api-url", "", "base url of pixie itself")

	grpcCmd.Flags().String("grpc-ca-cert", "", "Path to the grpc ca cert file")
	grpcCmd.Flags().String("grpc-cert", "", "Path to the grpc client cert file")
	grpcCmd.Flags().String("grpc-key", "", "Path to the grpc client key file")
	grpcCmd.Flags().String("grpc-address", "", "address of the grpc server")
	grpcCmd.Flags().String("metal-api-view-hmac", "", "hmac with metal-api view access")
	grpcCmd.Flags().String("metal-api-url", "", "url to access metal-api")
	grpcCmd.Flags().Bool("metal-hammer-debug", true, "set metal-hammer to debug")

	// metal-hammer remote logging configuration
	grpcCmd.Flags().String("metal-hammer-logging-endpoint", "", "set metal-hammer to send logs to this endpoint")
	grpcCmd.Flags().String("metal-hammer-logging-user", "", "set metal-hammer to send logs to a remote endpoint and authenticate with this user")
	grpcCmd.Flags().String("metal-hammer-logging-password", "", "set metal-hammer to send logs to a remote endpoint and authenticate with this password")
	grpcCmd.Flags().String("metal-hammer-logging-cert", "", "set metal-hammer to send logs to a remote endpoint and authenticate with this cert")
	grpcCmd.Flags().String("metal-hammer-logging-key", "", "set metal-hammer to send logs to a remote endpoint and authenticate with this key")
	grpcCmd.Flags().Bool("metal-hammer-logging-tls-insecure", false, "set metal-hammer to send logs to a remote endpoint without verifying the tls certificate")
	grpcCmd.Flags().String("metal-hammer-logging-type", "loki", "set metal-hammer to send logs to a remote endpoint with this logging type")
}

func getMetalAPIConfig(cmd *cobra.Command) (*api.MetalConfig, error) {
	grpcCACertFile, err := cmd.Flags().GetString("grpc-ca-cert")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	caCert, err := os.ReadFile(grpcCACertFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read ca-cert %w", err)
	}

	grpcClientCertFile, err := cmd.Flags().GetString("grpc-cert")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
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
		return nil, fmt.Errorf("error reading flag: %w", err)
	}

	hmac, err := cmd.Flags().GetString("metal-api-view-hmac")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalAPIUrl, err := cmd.Flags().GetString("metal-api-url")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	_, err = url.Parse(metalAPIUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse metal-api-url: %w", err)
	}
	pixieAPIUrl, err := cmd.Flags().GetString("pixie-api-url")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	_, err = url.Parse(pixieAPIUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pixie-api-url: %w", err)
	}
	metalHammerDebug, err := cmd.Flags().GetBool("metal-hammer-debug")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}

	// Log forwarding for the metal-hammer
	metalHammerLoggingEndpoint, err := cmd.Flags().GetString("metal-hammer-logging-endpoint")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingUser, err := cmd.Flags().GetString("metal-hammer-logging-user")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingPassword, err := cmd.Flags().GetString("metal-hammer-logging-password")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingCert, err := cmd.Flags().GetString("metal-hammer-logging-cert")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingKey, err := cmd.Flags().GetString("metal-hammer-logging-key")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingTlsInsecure, err := cmd.Flags().GetBool("metal-hammer-logging-tls-insecure")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	metalHammerLoggingType, err := cmd.Flags().GetString("metal-hammer-logging-type")
	if err != nil {
		return nil, fmt.Errorf("error reading flag: %w", err)
	}
	var logging *api.Logging
	if metalHammerLoggingEndpoint != "" {
		logging = &api.Logging{
			Endpoint: metalHammerLoggingEndpoint,
		}
		if metalHammerLoggingUser != "" {
			basicAuth := &api.BasicAuth{}
			basicAuth.User = metalHammerLoggingUser
			if metalHammerLoggingPassword != "" {
				basicAuth.Password = metalHammerLoggingUser
			}
			logging.BasicAuth = basicAuth
		}
		if metalHammerLoggingCert != "" && metalHammerLoggingKey != "" {
			cert, err := os.ReadFile(metalHammerLoggingCert)
			if err != nil {
				return nil, err
			}
			key, err := os.ReadFile(metalHammerLoggingKey)
			if err != nil {
				return nil, err
			}

			logging.CertificateAuth = &api.CertificateAuth{
				Cert:               string(cert),
				Key:                string(key),
				InsecureSkipVerify: metalHammerLoggingTlsInsecure,
			}
		}

		switch strings.ToLower(metalHammerLoggingType) {
		case "loki":
			logging.Type = api.LogTypeLoki
		default:
			return nil, fmt.Errorf("only loki currently support for metal-hammer remote logging %q was given", metalHammerLoggingType)
		}
	}

	return &api.MetalConfig{
		Debug:       metalHammerDebug,
		GRPCAddress: grpcAddress,
		MetalAPIUrl: metalAPIUrl,
		PixieAPIURL: pixieAPIUrl,
		CACert:      string(caCert),
		Cert:        string(clientCert),
		Key:         string(clientKey),
		HMAC:        hmac,
		Logging:     logging,
	}, nil
}
