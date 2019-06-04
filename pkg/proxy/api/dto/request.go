package dto

import (
	"encoding/base64"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"time"
)

// RequestDTO is DTO for request
type RequestDTO struct {
	ID           string    `json:"id"`
	Service      string    `json:"service"`
	RequestTime  time.Time `json:"request_time"`
	RequestSize  int64     `json:"request_size"`
	ResponseTime time.Time `json:"response_time"`
	ResponseSize int64     `json:"response_size"`
	Request      string    `json:"request_body,omitempty"`
	Response     string    `json:"response_body,omitempty"`
}

// RequestsToDTO converts slice of request to DTOs
func RequestsToDTO(reqs []domain.Request) []RequestDTO {
	result := make([]RequestDTO, len(reqs))
	for i := 0; i < len(reqs); i++ {
		result[i] = RequestToDTO(reqs[i])
	}
	return result
}

// RequestToDTO converts domain object to DTO without request/response body elements
func RequestToDTO(req domain.Request) RequestDTO {
	return RequestDTO{
		ID:           req.ID,
		Service:      req.Service,
		RequestTime:  req.RequestTime,
		ResponseTime: req.ResponseTime,
		RequestSize:  req.RequestSize,
		ResponseSize: req.ResponseSize,
	}
}

// RequestToFullDTO converts domain object to DTO with all fields included
func RequestToFullDTO(req domain.Request) RequestDTO {
	encoding := base64.StdEncoding
	return RequestDTO{
		ID:           req.ID,
		Service:      req.Service,
		RequestTime:  req.RequestTime,
		ResponseTime: req.ResponseTime,
		RequestSize:  req.RequestSize,
		ResponseSize: req.ResponseSize,
		Request:      encoding.EncodeToString(req.Request),
		Response:     encoding.EncodeToString(req.Response),
	}
}
