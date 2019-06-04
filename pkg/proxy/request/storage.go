package request

import (
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/bluele/gcache"
	"time"
)

const (
	defaultExpiration = 90 * time.Minute
	cacheMaxSize      = 100
)

// Storage is cache for request and their responses
type Storage interface {
	Set(req domain.Request)
	Get(ID string) (domain.Request, bool)
	GetAllIDs() []string
	GetAll() []domain.Request
}

type requestCache struct {
	cache gcache.Cache
}

// NewStorage creates new request cache instance
func NewStorage(storageSize int, storageDuration time.Duration) Storage {
	size := storageSize
	if storageSize <= 0 {
		size = cacheMaxSize
	}

	expiration := storageDuration
	if storageDuration <= 0 {
		expiration = defaultExpiration
	}

	gc := gcache.New(size).
		LRU().
		Expiration(expiration).
		Build()

	return &requestCache{
		cache: gc,
	}
}

func (c *requestCache) Set(req domain.Request) {
	_ = c.cache.Set(req.ID, req)
}

func (c *requestCache) Get(ID string) (domain.Request, bool) {
	item, err := c.cache.Get(ID)
	if err != nil {
		return domain.Request{}, false
	}

	req, ok := item.(domain.Request)
	if !ok {
		return domain.Request{}, false
	}
	return req, true
}

func (c *requestCache) GetAllIDs() []string {
	items := c.cache.Keys(false)

	result := make([]string, len(items))
	for i, key := range items {
		ID, _ := key.(string)
		result[i] = ID
		i++
	}

	return result
}

func (c *requestCache) GetAll() []domain.Request {
	items := c.cache.GetALL(false)

	result := make([]domain.Request, len(items))
	i := 0
	for _, value := range items {
		req, _ := value.(domain.Request)
		result[i] = req
		i++
	}

	return result
}
