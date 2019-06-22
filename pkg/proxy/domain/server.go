package domain

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/aldas/xroad-mock-proxy/pkg/config/common"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// ProxyServers is collections type for ProxyServer instances
type ProxyServers []ProxyServer

// ProxyServer is proxy server where request can be proxied
type ProxyServer struct {
	Name      string
	Address   url.URL
	IsDefault bool
	Transport http.RoundTripper
}

// ConvertProxyServers converts configuration to domain object
func ConvertProxyServers(conf config.ProxyServerConfigs) (ProxyServers, error) {
	result := ProxyServers{}
	for _, c := range conf {
		server, err := createProxyServer(c)
		if err != nil {
			return ProxyServers{}, err
		}
		result = append(result, server)
	}
	return result, nil
}

func createProxyServer(conf config.ProxyServerConf) (ProxyServer, error) {
	var transport = http.DefaultTransport

	if !conf.TLS.UseSystemTransport && conf.TLS.CertFile != "" {
		tlsClientConf, err := tlsConfig(conf.TLS)
		if err != nil {
			return ProxyServer{}, err
		}

		transport = &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsClientConf,
			MaxIdleConns:        100,
			MaxConnsPerHost:     25, // we should usually have very few different clients connecting
		}
	}

	address, err := url.Parse(conf.Address)
	if err != nil {
		return ProxyServer{}, errors.Wrap(err, "failed to parse proxy server address to url")
	}

	return ProxyServer{
		Name:      conf.Name,
		Address:   *address,
		IsDefault: conf.IsDefault,
		Transport: transport,
	}, nil
}

func tlsConfig(config common.TLSConf) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to proxy load key pair")
	}

	caCert, err := ioutil.ReadFile(config.CAFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read proxy CA cert")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

// Default finds default server
func (p ProxyServers) Default() (ProxyServer, bool) {
	for _, server := range p {
		if server.IsDefault {
			return server, true
		}
	}
	return ProxyServer{}, false
}

// Find returns first proxy server matching given name
func (p ProxyServers) Find(name string) (ProxyServer, bool) {
	for _, server := range p {
		if server.Name == name {
			return server, true
		}
	}
	return ProxyServer{}, false
}

// FindByHost returns first proxy server matching given host
func (p ProxyServers) FindByHost(host string) (ProxyServer, bool) {
	for _, server := range p {
		if server.Address.Host == host {
			return server, true
		}
	}
	return ProxyServer{}, false
}
