package config

import (
	"github.com/aldas/xroad-mock-proxy/pkg/mock/config"
	configProxy "github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
)

// Config is root element for application configuration
type Config struct {
	ProxyConf configProxy.ProxyConf `mapstructure:"proxy"`
	MockConf  config.MockConf       `mapstructure:"mock"`
}
