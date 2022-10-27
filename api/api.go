package api

type MetalConfig struct {
	Debug       bool
	GRPCAddress string `json:"address,omitempty"`
	MetalAPIUrl string `json:"metal_api_url,omitempty"`
	PixieAPIURL string `json:"pixie_api_url"`
	CACert      string `json:"ca_cert,omitempty"`
	Cert        string `json:"cert,omitempty"`
	Key         string `json:"key,omitempty"`
	HMAC        string `json:"hmac,omitempty"`
}
