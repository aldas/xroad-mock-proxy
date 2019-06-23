package domain

import (
	"crypto/tls"
	"github.com/aldas/xroad-mock-proxy/pkg/common/server"
	"github.com/aldas/xroad-mock-proxy/pkg/config/common"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProxyServers is collections type for ProxyServer instances
type ProxyServers []ProxyServer

// ProxyServer is proxy server where request can be proxied
type ProxyServer struct {
	ID         int64
	Name       string
	Address    url.URL
	IsDefault  bool
	IsReadOnly bool
	Transport  http.RoundTripper
}

// ConvertProxyServers converts configuration to domain object
func ConvertProxyServers(conf config.ProxyServerConfigs) (ProxyServers, error) {
	result := ProxyServers{}

	i := int64(1)
	for _, c := range conf {
		s, err := createProxyServer(c)
		if err != nil {
			return ProxyServers{}, err
		}
		s.ID = i
		i++

		result = append(result, s)
	}
	return result, nil
}

func createProxyServer(conf config.ProxyServerConf) (ProxyServer, error) {
	var transport = http.DefaultTransport

	if !conf.TLS.UseSystemTransport && conf.TLS.CertFile != "" {
		tlsClientConf, err := confToTLSConfig(conf.TLS)
		if err != nil {
			return ProxyServer{}, err
		}
		transport = ProxyServerTransport(tlsClientConf)
	}

	address, err := url.Parse(conf.Address)
	if err != nil {
		return ProxyServer{}, errors.Wrap(err, "failed to parse proxy server address to url")
	}

	isReadOnly := true
	if conf.IsReadOnly != nil {
		isReadOnly = *conf.IsReadOnly
	}

	return ProxyServer{
		Name:       strings.ToLower(conf.Name),
		Address:    *address,
		IsDefault:  conf.IsDefault,
		IsReadOnly: isReadOnly,
		Transport:  transport,
	}, nil
}

// ProxyServerTransport creates proxy server transport for default settings
func ProxyServerTransport(tlsClientConf *tls.Config) *http.Transport {
	return &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsClientConf,
		MaxIdleConns:        100,
		MaxConnsPerHost:     25, // we should usually have very few different clients connecting
	}
}

func confToTLSConfig(config common.TLSConf) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(config.CAFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read proxy CA cert")
	}

	certBytes, err := ioutil.ReadFile(config.CertFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cert file")
	}

	keyBytes, err := ioutil.ReadFile(config.KeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key file")
	}

	return server.ToTLSConfig(caCert, certBytes, keyBytes, config.KeyPassword)
}

// Default finds default server
func (p ProxyServers) Default() (ProxyServer, bool) {
	for _, s := range p {
		if s.IsDefault {
			return s, true
		}
	}
	return ProxyServer{}, false
}

// Find returns first proxy server matching given name
func (p ProxyServers) Find(name string) (ProxyServer, bool) {
	name = strings.ToLower(name)

	for _, s := range p {
		if s.Name == name {
			return s, true
		}
	}
	return ProxyServer{}, false
}

// FindByHost returns first proxy server matching given host
func (p ProxyServers) FindByHost(host string) (ProxyServer, bool) {
	for _, s := range p {
		// NB: Address.Host contains already port. ie. 'localhost:443'
		if s.Address.Host == host {
			return s, true
		}
	}
	return ProxyServer{}, false
}
