package server

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/rs/zerolog"
)

// Service provides methods for proxy servers
type Service interface {
	ManagerService
	AccessorService
}

// AccessorService provides methods to access proxy servers
type AccessorService interface {
	HostToProxyServer(host string) (domain.ProxyServer, bool)
	DefaultServer() (domain.ProxyServer, bool)
	Find(name string) (domain.ProxyServer, bool)
	Servers() domain.ProxyServers
}

// ManagerService provides methods to manage proxy servers
type ManagerService interface {
	Save(domain.ProxyServer) (domain.ProxyServer, error)
	Remove(ID int64) bool
}

type service struct {
	logger  *zerolog.Logger
	storage Storage
}

// NewService created new server service instance
func NewService(logger *zerolog.Logger, servers domain.ProxyServers) Service {
	storage := NewStorage(logger, servers, cacheMaxSize, cacheDefaultDuration)

	return &service{
		logger:  logger,
		storage: storage,
	}
}

func (s service) HostToProxyServer(host string) (domain.ProxyServer, bool) {
	return s.storage.GetAll().FindByHost(host)
}

func (s service) DefaultServer() (domain.ProxyServer, bool) {
	return s.storage.GetAll().Default()
}

func (s service) Find(name string) (domain.ProxyServer, bool) {
	return s.storage.GetAll().Find(name)
}

func (s service) Save(server domain.ProxyServer) (domain.ProxyServer, error) {
	return s.storage.Save(server)
}

func (s service) Remove(ID int64) bool {
	// TODO remove rules
	return s.storage.Remove(ID)
}

func (s service) Servers() domain.ProxyServers {
	return s.storage.GetAll()
}
