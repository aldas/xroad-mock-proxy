package common

// TLSConf describes TLS configuration for any server.
type TLSConf struct {
	UseSystemTransport bool   `mapstructure:"use_system_transport"`
	CAFile             string `mapstructure:"ca_file"`
	CertFile           string `mapstructure:"cert_file"`
	KeyFile            string `mapstructure:"key_file"`
	KeyPassword        string `mapstructure:"key_password"`
}

// ServerTLSConf describes TLS configuration application HTTPS server
type ServerTLSConf struct {
	UseSystemTransport  bool   `mapstructure:"use_system_transport"`
	CAFile              string `mapstructure:"ca_file"`
	CertFile            string `mapstructure:"cert_file"`
	KeyFile             string `mapstructure:"key_file"`
	KeyPassword         string `mapstructure:"key_password"`
	ForceClientCertAuth bool   `mapstructure:"force_client_cert_auth"`
}
