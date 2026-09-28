package api

// MetalConfig is consumed by metal-hammer to get all options to open a grpc connection to the metal-api
// Deprecated: This mixes up the pixie configuration with server responses, this will be removed once the
// metal-hammer was fully migrated to V2MetalHammerConfigPayload
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

	// metal-apiserver related fields
	MetalAPIServerUrl       string `json:"-"`
	MetalAPIServerTokenFile string `json:"-"`
	MetalHammerTenant       string `json:"-"`

	// Logging contains logging configurations passed to metal-hammer
	Logging *Logging `json:"logging,omitempty"`
}

// V2MetalHammerConfigPayload contains initial configuration options for the metal-hammer.
type V2MetalHammerConfigPayload struct {
	// Client contains data for initializing the v2 metal-apiserver client
	Client Client `json:"client"`

	// NTPServers to use for synchronize time in the metal-hammer, defaults back to public ntp servers if empty
	NTPServers []string `json:"ntp_servers,omitempty"`
	// Partition used for registering the machine in the metal-hammer
	Partition string `json:"partition"`

	// Logging contains logging configurations passed to metal-hammer
	Logging *Logging `json:"logging,omitempty"`
}

type Client struct {
	// ApiUrl contains the URL to the metal-apiserver
	ApiUrl string `json:"metal_apiserver_url"`
	// Token contains the token for the client
	Token string `json:"metal_hammer_token"`

	// CACert contains the ca for the client
	CACert *string `json:"ca_cert,omitempty"`
	// Cert contains the cert for the client
	Cert *string `json:"cert,omitempty"`
	// Key contains the cert key for the client
	Key *string `json:"key,omitempty"`
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

// LogType defines which logging backend should be used
type LogType string

const (
	// LogTypeLoki loki is the logging backend
	LogTypeLoki = LogType("loki")
)
