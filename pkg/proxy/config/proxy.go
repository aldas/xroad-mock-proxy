package config

import (
	"github.com/aldas/xroad-mock-proxy/pkg/config/common"
	"time"
)

// ProxyConf is root config structure for proxy
type ProxyConf struct {
	ServerConf  ServerConf        `mapstructure:"server"`
	APIConf     APIConf           `mapstructure:"api"`
	StorageConf ServerStorageConf `mapstructure:"storage"`
	RoutesConf  RoutesConf        `mapstructure:"routes"`
}

// ServerConf describes server configuration that application starts
type ServerConf struct {
	Enabled             bool                 `mapstructure:"enabled"`
	Address             string               `mapstructure:"address"`
	ContextPath         string               `mapstructure:"context_path"`
	TLS                 common.ServerTLSConf `mapstructure:"tls"`
	ReadTimeoutSeconds  int                  `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int                  `mapstructure:"write_timeout_seconds"`
	Debug               bool                 `mapstructure:"is_debug"`
}

// APIConf describes API server configuration that application starts
type APIConf struct {
	Enabled             bool                 `mapstructure:"enabled"`
	Address             string               `mapstructure:"address"`
	ContextPath         string               `mapstructure:"context_path"`
	AssetsDirectory     string               `mapstructure:"assets_directory"`
	TLS                 common.ServerTLSConf `mapstructure:"tls"`
	ReadTimeoutSeconds  int                  `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int                  `mapstructure:"write_timeout_seconds"`
	Debug               bool                 `mapstructure:"is_debug"`
	DebugPath           string               `mapstructure:"debug_path"`
}

// RoutesConf describes configuration for x-road proxy. Contains settings for all servers and rules how to route/proxy requests
type RoutesConf struct {
	Servers ProxyServerConfigs `mapstructure:"servers"`
	Rules   RuleConfigs        `mapstructure:"rules"`
}

// ProxyServerConfigs is collections type for ProxyServerConf structure
type ProxyServerConfigs []ProxyServerConf

// ProxyServerConf is proxy server configurations where request can be proxied
type ProxyServerConf struct {
	Address   string         `mapstructure:"address"`
	TLS       common.TLSConf `mapstructure:"tls"`
	Name      string         `mapstructure:"name"`
	IsDefault bool           `mapstructure:"is_default"`
	// should rule be changeable in API (defaults to true)
	IsReadOnly *bool `mapstructure:"read_only"`
}

// ServerStorageConf describes serve different storage configurations
type ServerStorageConf struct {
	Requests StorageConf `mapstructure:"requests"`
	Rules    StorageConf `mapstructure:"rules"`
}

// StorageConf describes storage configuration
type StorageConf struct {
	Size       int           `mapstructure:"size"`
	Expiration time.Duration `mapstructure:"expiration"`
}
