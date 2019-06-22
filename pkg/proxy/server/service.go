package server

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/rs/zerolog"
)

// Service provides methods to manager proxy servers
type Service interface {
	HostToProxyServer(host string) (domain.ProxyServer, bool)
	DefaultServer() (domain.ProxyServer, bool)
	Servers() domain.ProxyServers
}

type service struct {
	logger  *zerolog.Logger
	servers domain.ProxyServers
}

// NewService created new server service instance
func NewService(logger *zerolog.Logger, servers domain.ProxyServers) Service {
	return &service{
		logger:  logger,
		servers: servers,
	}
}

func (s service) HostToProxyServer(host string) (domain.ProxyServer, bool) {
	return s.servers.FindByHost(host)
}

func (s service) DefaultServer() (domain.ProxyServer, bool) {
	return s.servers.Default()
}

func (s service) Servers() domain.ProxyServers {
	return s.servers // TODO: mutex & return copy
}
