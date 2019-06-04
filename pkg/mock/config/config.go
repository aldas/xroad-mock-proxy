package config

import (
	"github.com/aldas/xroad-mock-proxy/pkg/config/common"
)

// MockConf describes X-road mock server configuration
type MockConf struct {
	Enabled             bool                 `mapstructure:"enabled"`
	Address             string               `mapstructure:"address"`
	ContextPath         string               `mapstructure:"context_path"`
	TLS                 common.ServerTLSConf `mapstructure:"tls"`
	ReadTimeoutSeconds  int                  `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int                  `mapstructure:"write_timeout_seconds"`
	Debug               bool                 `mapstructure:"is_debug"`
	DebugPath           string               `mapstructure:"debug_path"`
	Rules               RuleConfigs          `mapstructure:"rules"`
	WebAssetsDirectory  string               `mapstructure:"web_assets_directory"`
	Storage             StorageConf          `mapstructure:"storage"`
}

// StorageConf describes rules storage configuration
type StorageConf struct {
	Size int `mapstructure:"size"`
}
