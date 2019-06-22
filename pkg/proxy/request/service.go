package request

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/apperror"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/rs/zerolog"
)

// Service is service for requests domain
type Service interface {
	DeleteRequests()
	GetRequests() []domain.Request
	GetRequest(ID string) (domain.Request, error)
}

type service struct {
	logger *zerolog.Logger
	cache  Storage
}

// NewService creates new instance of request service
func NewService(logger *zerolog.Logger, cache Storage) Service {
	return &service{
		logger: logger,
		cache:  cache,
	}
}

func (s service) DeleteRequests() {
	s.cache.DeleteAll()
}

func (s service) GetRequests() []domain.Request {
	return s.cache.GetAll()
}

func (s service) GetRequest(ID string) (domain.Request, error) {
	req, ok := s.cache.Get(ID)
	if !ok {
		return domain.Request{}, apperror.ErrorNotFound
	}

	return req, nil
}
