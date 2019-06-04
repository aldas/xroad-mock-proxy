package proxy

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/rs/zerolog"
)

// Service provides functionality to proxy handler for matching rules and servers
type Service interface {
	HostToProxyServer(host string) (domain.ProxyServer, bool)
	DefaultServer() (domain.ProxyServer, bool)
	Rules() domain.Rules
	Servers() domain.ProxyServers
}

type service struct {
	logger      *zerolog.Logger
	servers     domain.ProxyServers
	ruleService rule.Service
}

// NewService created new proxy service instance
func NewService(logger *zerolog.Logger, servers domain.ProxyServers, ruleService rule.Service) Service {
	return &service{
		logger:      logger,
		servers:     servers,
		ruleService: ruleService,
	}
}

func (s service) HostToProxyServer(host string) (domain.ProxyServer, bool) {
	return s.servers.FindByHost(host)
}

func (s service) DefaultServer() (domain.ProxyServer, bool) {
	return s.servers.Default()
}

func (s service) Rules() domain.Rules {
	return s.ruleService.GetAll()
}

func (s service) Servers() domain.ProxyServers {
	return s.servers // TODO: mutex & return copy
}
