package domain

import "time"

// Request is cached bodies of proxied request/response
type Request struct {
	ID           string
	RuleID       int64
	Service      string
	Request      []byte
	RequestTime  time.Time
	RequestSize  int64
	Response     []byte
	ResponseTime time.Time
	ResponseSize int64
}
