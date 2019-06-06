package rule

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/apperror"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"math/rand"
)

const (
	cacheMaxSize = 200

	// see: https://stackoverflow.com/a/307200 size: (1<<53)-1
	maxJavascriptInteger = 9007199254740991
)

// Storage provides storage for rules
type Storage interface {
	StorageGetter
	Save(domain.Rule) (domain.Rule, error)
	Remove(ID int64) bool
}

// StorageGetter provides interface to get rules from storage
type StorageGetter interface {
	GetAll() domain.Rules
	GetRule(ID int64) (domain.Rule, bool)
}

type cacheStorage struct {
	logger *zerolog.Logger
	rules  domain.Rules
	cache  gcache.Cache
}

// NewStorage creates new instance of rule storage using cache as storage
func NewStorage(logger *zerolog.Logger, rules domain.Rules, storageSize int) Storage {
	size := storageSize
	if storageSize <= 0 {
		size = cacheMaxSize
	}

	gc := gcache.New(size).
		LRU().
		Build()

	return &cacheStorage{
		logger: logger,
		rules:  rules,
		cache:  gc,
	}
}

func (s cacheStorage) GetAll() domain.Rules {
	raw := s.cache.GetALL(false)

	rules := make(domain.Rules, len(raw))
	i := 0
	for _, rawRule := range raw {
		rule, _ := rawRule.(domain.Rule)
		rules[i] = rule
		i++
	}

	return append(rules, s.rules...)
}

func (s cacheStorage) getRule(ID int64) (domain.Rule, bool) {
	// assuming we do not have huge collection of rules from config looping over slice
	// should be fast enough
	for _, rule := range s.rules {
		if rule.ID == ID {
			return rule, true
		}
	}
	return domain.Rule{}, false
}

func (s cacheStorage) GetRule(ID int64) (domain.Rule, bool) {
	rule, ok := s.getRule(ID)
	if ok {
		return rule, ok
	}
	return s.cacheGet(ID)
}

func (s cacheStorage) cacheGet(ID int64) (domain.Rule, bool) {
	raw, err := s.cache.Get(ID)
	if err != nil {
		return domain.Rule{}, false
	}

	rule, ok := raw.(domain.Rule)
	if !ok {
		return domain.Rule{}, false
	}
	return rule, true
}

func (s cacheStorage) Save(r domain.Rule) (domain.Rule, error) {
	var rule domain.Rule
	var ok bool

	if r.ID != 0 {
		rule, ok = s.getRule(r.ID)
		if ok {
			return domain.Rule{}, errors.New("can not change system rule")
		}

		rule, ok = s.getRule(r.ID)
		if !ok {
			return domain.Rule{}, apperror.ErrorNotFound
		}
	} else {
		// yeah, we could get already existing one
		r.ID = rand.Int63n(maxJavascriptInteger)
	}

	if ok && rule.IsReadOnly {
		return rule, errors.New("can not modify read only rule")
	}

	if err := s.cache.Set(r.ID, r); err != nil {
		return r, errors.Wrap(err, "failed to store rule in cache")
	}

	return r, nil
}

func (s cacheStorage) Remove(ID int64) bool {
	return s.cache.Remove(ID)
}
