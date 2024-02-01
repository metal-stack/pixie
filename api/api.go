package api

// MetalConfig is consumed by metal-hammer to get all options to open a grpc connection to the metal-api
type MetalConfig struct {
	Debug       bool   `json:"debug"`
	GRPCAddress string `json:"address,omitempty"`
	MetalAPIUrl string `json:"metal_api_url,omitempty"`
	PixieAPIURL string `json:"pixie_api_url"`
	CACert      string `json:"ca_cert,omitempty"`
	Cert        string `json:"cert,omitempty"`
	Key         string `json:"key,omitempty"`
	HMAC        string `json:"hmac,omitempty"`
	// Logging contains logging configurations passed to metal-hammer
	Logging *Logging `json:"logging,omitempty"`
}

type Logging struct {
	// Endpoint is the url where the logs must be shipped to
	Endpoint        string           `json:"endpoint,omitempty"`
	BasicAuth       *BasicAuth       `json:"basic_auth,omitempty"`
	CertificateAuth *CertificateAuth `json:"certificate_auth,omitempty"`
	// Type of logging
	Type LogType `json:"log_type,omitempty"`
}

type BasicAuth struct {
	// User to authenticate against the logging endpoint
	User string `json:"user,omitempty"`
	// Password to authenticate against the logging endpoint
	Password string `json:"password,omitempty"`
}
type CertificateAuth struct {
	Cert               string `json:"cert,omitempty"`
	Key                string `json:"key,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
}

type LogType string

const (
	LogTypeLoki = LogType("loki")
)
