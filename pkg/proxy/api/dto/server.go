package dto

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/server"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

// ServerDTO is DTO for proxyServer
type ServerDTO struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	IsReadOnly bool    `json:"read_only"`
	TLS        *TLSDTO `json:"tls,omitempty"`
}

// TLSDTO is DTO for server TLS configuration
type TLSDTO struct {
	UseSystemTransport bool   `json:"use_system_transport"`
	CACert             string `json:"ca_cert"`
	Cert               string `json:"cert"`
	Key                string `json:"key"`
	KeyPassword        string `json:"key_password"`
}

// ProxyServersToDTO converts slice of servers to DTOs
func ProxyServersToDTO(rules []domain.ProxyServer) []ServerDTO {
	result := make([]ServerDTO, len(rules))
	for i := 0; i < len(rules); i++ {
		result[i] = ProxyServerToDTO(rules[i])
	}
	return result
}

// ProxyServerToDTO converts proxy server to DTO
func ProxyServerToDTO(s domain.ProxyServer) ServerDTO {
	return ServerDTO{
		ID:         s.ID,
		Name:       s.Name,
		Address:    s.Address.String(),
		IsReadOnly: s.IsReadOnly,
	}
}

// ToProxyServer converts DTO object to proxyServer domain object
func ToProxyServer(s ServerDTO) (domain.ProxyServer, error) {
	address, err := url.Parse(s.Address)
	if err != nil {
		return domain.ProxyServer{}, errors.Wrap(err, "failed to parse address to url")
	}

	caCertBytes := []byte(s.TLS.CACert)
	certBytes := []byte(s.TLS.Cert)
	keyBytes := []byte(s.TLS.Key)
	tls, err := server.ToTLSConfig(caCertBytes, certBytes, keyBytes, s.TLS.KeyPassword)
	if err != nil {
		return domain.ProxyServer{}, err
	}

	return domain.ProxyServer{
		ID:         0,
		Name:       strings.ToLower(s.Name),
		Address:    *address,
		IsDefault:  false,
		IsReadOnly: s.IsReadOnly,
		Transport:  domain.ProxyServerTransport(tls),
	}, nil
}
