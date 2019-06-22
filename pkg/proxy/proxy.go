package proxy

import (
	"bytes"
	"fmt"
	"github.com/aldas/xroad-mock-proxy/pkg/common/soap"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/server"
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

type proxy struct {
	logger *zerolog.Logger
	cache  request.Storage

	serverService server.Service
	ruleService   rule.Service

	defaultServer domain.ProxyServer
	proxyHandler  http.Handler
}

func (p proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.proxyHandler.ServeHTTP(rw, req)
}

// NewProxyHandler creates new http handler for proxy
func NewProxyHandler(
	logger *zerolog.Logger,
	serverService server.Service,
	ruleService rule.Service,
	cache request.Storage,
) (http.Handler, error) {
	defaultServer, ok := serverService.DefaultServer()
	if !ok {
		return nil, errors.New("failed to find default proxy server configuration")
	}

	proxy := proxy{
		logger:        logger,
		serverService: serverService,
		ruleService:   ruleService,
		cache:         cache,

		defaultServer: defaultServer,
	}
	proxy.proxyHandler = proxy.createProxyHandler()

	return proxy, nil
}

func (p *proxy) createProxyHandler() http.Handler {
	defaultProxyURL := p.defaultServer.Address

	director := func(req *http.Request) {
		proxyURL := defaultProxyURL
		if req.Body != nil {
			if tmpURL := p.processBody(req); tmpURL != nil {
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
		logger:        p.logger,
		serverService: p.serverService,
	}
	if p.defaultServer.Transport != nil {
		switcher.Transport = p.defaultServer.Transport
	}
	proxy.Transport = switcher

	proxy.ModifyResponse = p.modifyResponse

	return proxy
}

func (p *proxy) processBody(req *http.Request) *url.URL {
	// read all bytes from content body and create new stream using it.
	requestBody, _ := ioutil.ReadAll(req.Body)

	// TODO: handle multipart requests - detect from headers?
	// TODO: "Content-Type: Multipart/Related" https://www.w3.org/TR/SOAP-attachments
	soapService, err := soap.FromRequestBody(requestBody)
	if err != nil {
		// let request through if we can not handle it. it will go to default server
		req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		p.logger.Error().Err(err).Msg("unable to extract service info from request")
		return nil
	}
	serviceName := soapService.Service

	logRow := p.logger.Info().Str("serviceName", serviceName)

	// TODO match Request.Header
	matchedRule, ok := p.ruleService.GetAll().MatchRemoteAddr(req.RemoteAddr).MatchService(serviceName).MatchRegex(requestBody)
	if !ok {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		logRow.Msg("received SOAP message without matching rule")
		return nil
	}

	requestID := fmt.Sprintf("%v", rand.Uint64())
	p.cache.Set(domain.Request{
		ID:          requestID,
		RuleID:      matchedRule.ID,
		Service:     serviceName,
		RequestTime: time.Now(),
		Request:     requestBody,
		RequestSize: int64(len(requestBody)),
	})
	req.Header.Add(requestIDHeader, requestID)
	// ruleID is also in header because by the time response arrives our LRU cache can be already dropped request
	// object but we need rule to response replacements to work
	req.Header.Add(requestRuleIDHeader, strconv.Itoa(int(matchedRule.ID)))

	logRow.Str("requestID", requestID).Int64("ruleID", matchedRule.ID).Msg("Matched to rule")

	matchedServer, ok := p.serverService.Servers().Find(matchedRule.Server)
	if !ok {
		p.logger.Error().Msg("failed to find server matching rule")

		req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		return nil
	}

	if len(matchedRule.RequestReplacements) > 0 {
		requestBody = matchedRule.ApplyRequestReplacements(requestBody)
		requestSize := int64(len(requestBody))

		req.ContentLength = requestSize
		req.Header.Set("Content-Length", strconv.Itoa(int(requestSize)))
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
	return &matchedServer.Address
}

func (p *proxy) modifyResponse(r *http.Response) error {
	requestID := r.Request.Header.Get(requestIDHeader)
	ruleIDStr := r.Request.Header.Get(requestRuleIDHeader)
	if ruleIDStr == "" && requestID == "" {
		return nil
	}

	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		p.logger.Error().Err(err).Str("rule_id", ruleIDStr).Msg("failed to convert rule id to int")
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
		r, ok := p.ruleService.GetAll().FindByID(int64(ruleID))
		if ok && len(r.ResponseReplacements) > 0 {
			responseBody = r.ApplyResponseReplacements(responseBody)
		}
	}

	responseSize := int64(len(responseBody))
	r.ContentLength = responseSize
	r.Header.Set("Content-Length", strconv.Itoa(int(responseSize)))

	r.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

	if requestID != "" {
		cached, ok := p.cache.Get(requestID)
		if ok {
			cached.Response = responseBody
			cached.ResponseTime = time.Now()
			cached.ResponseSize = responseSize
			p.cache.Set(cached)
		}
	}
	return nil
}
