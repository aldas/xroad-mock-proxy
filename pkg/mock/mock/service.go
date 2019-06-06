package mock

import (
	"bytes"
	"github.com/aldas/xroad-mock-proxy/pkg/common/soap"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/rule"
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

// Service provides mock functionality
type Service interface {
	mock(requestBody []byte) ([]byte, int)
}

type service struct {
	logger  *zerolog.Logger
	storage rule.StorageGetter
}

// NewService creates instance of mock service
func NewService(logger *zerolog.Logger, storage rule.StorageGetter) Service {
	return &service{
		logger:  logger,
		storage: storage,
	}
}

func (s service) mock(requestBody []byte) ([]byte, int) {
	soapService, err := soap.FromRequestBody(requestBody)
	if err != nil {
		// TODO: handle multipart requests - detect from headers?
		// TODO: "Content-Type: Multipart/Related" https://www.w3.org/TR/SOAP-attachments
		s.logger.Error().Err(err).Msg("failed to unmarshal request data")
		return []byte("Internal server error\n"), http.StatusInternalServerError
	}
	s.logger.Info().Str("service", soapService.Service).Msg("SOAP")

	matchedRule, ok := s.storage.GetAll().MatchService(soapService.Service).MatchRegex(requestBody)
	if !ok {
		return []byte("Rule not found\n"), http.StatusNotFound
	}
	s.logger.Debug().
		Str("service", soapService.Service).
		Int64("rule_id", matchedRule.ID).
		Msg("serving mock response")

	identity, ok := matchedRule.MatchIdentity(requestBody)
	if !ok {
		return []byte("Unable to find identity in request\n"), http.StatusNotFound
	}

	return s.processRule(matchedRule, identity)
}

func (s service) processRule(matchedRule domain.Rule, identity string) ([]byte, int) {
	vars := fromIdentity(identity)

	var tpl bytes.Buffer
	err := matchedRule.Template.Execute(&tpl, vars)
	if err != nil {
		s.logger.Error().Err(err).Interface("vars", vars).Msg("failed to execute template")
		return []byte("Internal server error\n"), http.StatusInternalServerError
	}

	if matchedRule.Timeout != 0 {
		time.Sleep(matchedRule.Timeout)
	}

	return tpl.Bytes(), matchedRule.ResponseStatus
}
