package rule

import (
	"errors"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/rs/zerolog"
)

// Service provides mock API functionality
type Service interface {
	GetAll() domain.Rules
	GetRule(ID int64) (domain.Rule, bool)
	Save(domain.Rule) (domain.Rule, error)
	Remove(ID int64) bool
}

type service struct {
	logger  *zerolog.Logger
	storage Storage
}

// NewService creates instance of rule service
func NewService(logger *zerolog.Logger, storage Storage) Service {
	return &service{
		logger:  logger,
		storage: storage,
	}
}

func (s service) GetAll() domain.Rules {
	return s.storage.GetAll()
}

func (s service) GetRule(ID int64) (domain.Rule, bool) {
	return s.storage.GetRule(ID)
}

func (s service) Save(rule domain.Rule) (domain.Rule, error) {
	if rule.ID != 0 {
		tmp, ok := s.GetRule(rule.ID)
		if !ok {
			return rule, errors.New("trying to modify not existent rule")
		}
		if ok && tmp.IsReadOnly {
			return rule, errors.New("can not modify read only rule")
		}
	}

	return s.storage.Save(rule)
}

func (s service) Remove(ID int64) bool {
	tmp, ok := s.GetRule(ID)
	if !ok || tmp.IsReadOnly {
		return false
	}

	return s.storage.Remove(ID)
}
