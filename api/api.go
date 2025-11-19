package api

// MetalConfig is consumed by metal-hammer to get all options to open a grpc connection to the metal-api
type MetalConfig struct {
	Debug       bool     `json:"debug"`
	GRPCAddress string   `json:"address,omitempty"`
	MetalAPIUrl string   `json:"metal_api_url,omitempty"`
	PixieAPIURL string   `json:"pixie_api_url"`
	CACert      string   `json:"ca_cert,omitempty"`
	Cert        string   `json:"cert,omitempty"`
	Key         string   `json:"key,omitempty"`
	HMAC        string   `json:"hmac,omitempty"`
	NTPServers  []string `json:"ntp_servers,omitempty"`
	Partition   string   `json:"partition"`
	// Logging contains logging configurations passed to metal-hammer
	Logging   *Logging     `json:"logging,omitempty"`
	OciConfig []*OciConfig `json:"oci_config,omitempty"`
}

type Logging struct {
	// Endpoint is the url where the logs must be shipped to
	Endpoint string `json:"endpoint,omitempty"`
	// BasicAuth must be set if loki requires username and password
	BasicAuth *BasicAuth `json:"basic_auth,omitempty"`
	// CertificateAuth must be set if mTLS authentication is required for loki
	CertificateAuth *CertificateAuth `json:"certificate_auth,omitempty"`
	// Type of logging
	Type LogType `json:"log_type,omitempty"`
}

// BasicAuth configuration
type BasicAuth struct {
	// User to authenticate against the logging endpoint
	User string `json:"user,omitempty"`
	// Password to authenticate against the logging endpoint
	Password string `json:"password,omitempty"`
}

// CertificateAuth is used for mTLS authentication
type CertificateAuth struct {
	// Cert the certificate
	Cert string `json:"cert,omitempty"`
	// Key is the key
	Key string `json:"key,omitempty"`
	// InsecureSkipVerify if no certificate validation should be made
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`
}

type OciConfig struct {
	// URL pointing to the oci registry
	RegistryURL string `json:"registry_url"`
	// Username that is capable of logging in to the registry
	Username string `json:"username,omitempty"`
	// Password for the user
	Password string `json:"password,omitempty"`
}

// LogType defines which logging backend should be used
type LogType string

const (
	// LogTypeLoki loki is the logging backend
	LogTypeLoki = LogType("loki")
)
