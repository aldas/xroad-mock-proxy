package rule

import (
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/rs/zerolog"
)

// Service provides mock API functionality
type Service interface {
	GetAll() []domain.Rule
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

func (s service) GetAll() []domain.Rule {
	return s.storage.GetAll()
}

func (s service) GetRule(ID int64) (domain.Rule, bool) {
	return s.storage.GetRule(ID)
}

func (s service) Save(rule domain.Rule) (domain.Rule, error) {
	return s.storage.Save(rule)
}

func (s service) Remove(ID int64) bool {
	tmp, ok := s.GetRule(ID)
	if !ok || tmp.IsReadOnly {
		return false
	}

	return s.storage.Remove(ID)
}
