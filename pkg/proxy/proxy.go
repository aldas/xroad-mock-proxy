package proxy

import (
	"bytes"
	"fmt"
	"github.com/aldas/xroad-mock-proxy/pkg/common/soap"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

const (
	// XroadDefaulURL is default path where x-road is usually served from security server
	XroadDefaulURL string = "/cgi-bin/consumer_proxy"
	// requestIDHeader helps to find request in later stages handling responses
	requestIDHeader string = "X-Xroad-Proxy-Request-ID"
	// requestRuleIDHeader helps to find request rule in later stages to replace response contents
	requestRuleIDHeader string = "X-Xroad-Proxy-Rule-ID"
)

// NewProxyHandler creates new http handler for proxy
func NewProxyHandler(logger *zerolog.Logger, service Service, cache request.Storage) (http.Handler, error) {
	return proxyHandler(logger, service, cache)
}

func proxyHandler(logger *zerolog.Logger, service Service, cache request.Storage) (http.Handler, error) {
	defaultProxy, ok := service.DefaultServer()
	if !ok {
		return nil, errors.New("failed to find default proxy server configuration")
	}

	defaultProxyURL := defaultProxy.Address

	director := func(req *http.Request) {
		proxyURL := defaultProxyURL
		if req.Body != nil {
			if tmpURL := processBody(logger, service, cache, req); tmpURL != nil {
				proxyURL = *tmpURL
			}
		}

		// Host header needs to be changed to match our target server hostname otherwise target http server does
		// not understand that request is mean for it
		req.Host = proxyURL.Host

		// host/scheme where proxied request actually is sent
		req.URL.Host = proxyURL.Host
		req.URL.Scheme = proxyURL.Scheme
	}

	proxy := &httputil.ReverseProxy{
		Director:      director,
		FlushInterval: 1 * time.Second,
	}

	switcher := transportSwitcher{
		logger:  logger,
		service: service,
	}
	if defaultProxy.Transport != nil {
		switcher.Transport = defaultProxy.Transport
	}
	proxy.Transport = switcher

	proxy.ModifyResponse = func(r *http.Response) error {
		requestID := r.Request.Header.Get(requestIDHeader)
		ruleIDStr := r.Request.Header.Get(requestRuleIDHeader)
		if ruleIDStr == "" && requestID == "" {
			return nil
		}

		ruleID, err := strconv.Atoi(ruleIDStr)
		if err != nil {
			logger.Error().Err(err).Str("rule_id", ruleIDStr).Msg("failed to convert rule id to int")
			return nil
		}

		responseBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		err = r.Body.Close()
		if err != nil {
			return err
		}

		if ruleID != 0 {
			rule, ok := service.Rules().FindByID(int64(ruleID))
			if ok && len(rule.ResponseReplacements) > 0 {
				responseBody = rule.ApplyResponseReplacements(responseBody)
			}
		}

		responseSize := int64(len(responseBody))
		r.ContentLength = responseSize
		r.Header.Set("Content-Length", strconv.Itoa(int(responseSize)))

		r.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

		if requestID != "" {
			cached, ok := cache.Get(requestID)
			if ok {
				cached.Response = responseBody
				cached.ResponseTime = time.Now()
				cached.ResponseSize = responseSize
				cache.Set(cached)
			}
		}
		return nil
	}

	return proxy, nil
}

func processBody(logger *zerolog.Logger, service Service, cache request.Storage, req *http.Request) *url.URL {
	// read all bytes from content body and create new stream using it.
	requestBody, _ := ioutil.ReadAll(req.Body)

	var serviceName = ""
	if soapService, err := soap.FromRequestBody(requestBody); err == nil {
		// TODO: handle multipart requests - detect from headers?
		// TODO: "Content-Type: Multipart/Related" https://www.w3.org/TR/SOAP-attachments
		serviceName = soapService.Service

		logger.Info().
			Str("SubsystemCode", soapService.SubsystemCode).
			Str("ServiceCode", soapService.ServiceCode).
			Str("ServiceVersion", soapService.ServiceVersion).
			Msg("received SOAP message")
	}

	// TODO match Request.Header
	rule, ok := service.Rules().MatchRemoteAddr(req.RemoteAddr).MatchService(serviceName).MatchRegex(requestBody)
	if !ok {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		return nil
	}

	requestID := fmt.Sprintf("%v", rand.Uint64())
	cache.Set(domain.Request{
		ID:          requestID,
		RuleID:      rule.ID,
		Service:     serviceName,
		RequestTime: time.Now(),
		Request:     requestBody,
		RequestSize: int64(len(requestBody)),
	})
	req.Header.Add(requestIDHeader, requestID)
	// ruleID is also in header because by the time response arrives our LRU cache can be already dropped request
	// object but we need rule to response replacements to work
	req.Header.Add(requestRuleIDHeader, strconv.Itoa(int(rule.ID)))

	logger.Debug().
		Str("requestID", requestID).
		Int64("ruleID", rule.ID).
		Msg("Matched to rule")

	server, ok := service.Servers().Find(rule.Server)
	if !ok {
		logger.Error().Msg("failed to find server matching rule")

		req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		return nil
	}

	if len(rule.RequestReplacements) > 0 {
		requestBody = rule.ApplyRequestReplacements(requestBody)
		requestSize := int64(len(requestBody))

		req.ContentLength = requestSize
		req.Header.Set("Content-Length", strconv.Itoa(int(requestSize)))
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	return &server.Address
}
