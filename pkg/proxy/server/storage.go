package server

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/apperror"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"math/rand"
	"time"
)

const (
	cacheMaxSize         = 100
	cacheDefaultDuration = time.Duration(0)

	// see: https://stackoverflow.com/a/307200 size: (1<<53)-1
	maxJavascriptInteger = 9007199254740991
)

// Storage provides storage for servers
type Storage interface {
	GetAll() domain.ProxyServers
	GetServer(ID int64) (domain.ProxyServer, bool)
	Save(domain.ProxyServer) (domain.ProxyServer, error)
	Remove(ID int64) bool
}

type cacheStorage struct {
	logger  *zerolog.Logger
	servers domain.ProxyServers
	cache   gcache.Cache
}

// NewStorage creates new instance of server storage using in memory cache as storage backend
func NewStorage(logger *zerolog.Logger, servers domain.ProxyServers, storageSize int, storageDuration time.Duration) Storage {
	size := storageSize
	if storageSize <= 0 {
		size = cacheMaxSize
	}

	gcBuilder := gcache.New(size).LRU()

	if storageDuration > 0 {
		gcBuilder = gcBuilder.Expiration(storageDuration)
	}

	gc := gcBuilder.Build()

	return &cacheStorage{
		logger:  logger,
		servers: servers,
		cache:   gc,
	}
}

func (s cacheStorage) GetAll() domain.ProxyServers {
	raw := s.cache.GetALL(false)

	servers := make(domain.ProxyServers, len(raw))
	i := 0
	for _, rawServer := range raw {
		server, _ := rawServer.(domain.ProxyServer)
		servers[i] = server
		i++
	}

	return append(servers, s.servers...)
}

func (s cacheStorage) getServer(ID int64) (domain.ProxyServer, bool) {
	// assuming we do not have huge collection of server from config looping over slice
	// should be fast enough
	for _, server := range s.servers {
		if server.ID == ID {
			return server, true
		}
	}
	return domain.ProxyServer{}, false
}

func (s cacheStorage) GetServer(ID int64) (domain.ProxyServer, bool) {
	server, ok := s.getServer(ID)
	if ok {
		return server, ok
	}
	return s.cacheGet(ID)
}

func (s cacheStorage) cacheGet(ID int64) (domain.ProxyServer, bool) {
	raw, err := s.cache.Get(ID)
	if err != nil {
		return domain.ProxyServer{}, false
	}

	server, ok := raw.(domain.ProxyServer)
	if !ok {
		return domain.ProxyServer{}, false
	}
	return server, true
}

func (s cacheStorage) Save(r domain.ProxyServer) (domain.ProxyServer, error) {
	var server domain.ProxyServer
	var isExisting bool

	if r.ID != 0 {
		server, isExisting = s.getServer(r.ID)
		if isExisting {
			return domain.ProxyServer{}, errors.New("can not change system server")
		}

		server, isExisting = s.cacheGet(r.ID)
		if !isExisting {
			return domain.ProxyServer{}, apperror.ErrorNotFound
		}
	} else {
		// yeah, we could get already existing one
		r.ID = rand.Int63n(maxJavascriptInteger)
	}

	if isExisting && server.IsReadOnly {
		return server, errors.New("can not modify read only server")
	}

	if !s.isUniqueName(r) {
		return server, errors.New("invalid name. server name must be unique")
	}

	if err := s.cache.Set(r.ID, r); err != nil {
		return r, errors.Wrap(err, "failed to store server in cache")
	}

	return r, nil
}

func (s cacheStorage) isUniqueName(server domain.ProxyServer) bool {
	for _, srv := range s.GetAll() {
		if srv.Name == server.Name && srv.ID != server.ID {
			return false
		}
	}
	return true
}

func (s cacheStorage) Remove(ID int64) bool {
	return s.cache.Remove(ID)
}
