package proxy

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/server"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/labstack/echo"
	"github.com/rs/zerolog"
	"net/http"
	"sync"
)

//Start starts proxy server in goroutine
func Start(
	logger *zerolog.Logger,
	serverConfig config.ServerConf,
	proxyServerConfigs config.ProxyServerConfigs,
	requestCache request.Storage,
	ruleService rule.Service,
	wg *sync.WaitGroup,
) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := serve(logger, serverConfig, proxyServerConfigs, requestCache, ruleService)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to start proxy server")
		}
	}()
}

// Serve starts http(s) server. Method will block until server is shutdown gracefully or until context timeouts
func serve(
	logger *zerolog.Logger,
	serverConfig config.ServerConf,
	proxyServerConfigs config.ProxyServerConfigs,
	requestCache request.Storage,
	ruleService rule.Service,
) error {
	e := server.New()

	e.GET("/", echo.WrapHandler(defaultHandler{logger: logger}))

	proxyHandler, err := createProxyHandler(logger, proxyServerConfigs, requestCache, ruleService)
	if err != nil {
		return err
	}

	contextPath := serverConfig.ContextPath
	if contextPath == "" {
		contextPath = XroadDefaulURL
	}
	e.POST(contextPath, echo.WrapHandler(proxyHandler))

	err = server.Start(e, &server.Config{
		Address:             serverConfig.Address,
		ReadTimeoutSeconds:  serverConfig.ReadTimeoutSeconds,
		WriteTimeoutSeconds: serverConfig.WriteTimeoutSeconds,
		Debug:               serverConfig.Debug,
		TLS: server.TLSConf{
			CAFile:              serverConfig.TLS.CAFile,
			CertFile:            serverConfig.TLS.CertFile,
			KeyFile:             serverConfig.TLS.KeyFile,
			KeyPassword:         serverConfig.TLS.KeyPassword,
			ForceClientCertAuth: serverConfig.TLS.ForceClientCertAuth,
		},
	})

	return err
}

func createProxyHandler(
	logger *zerolog.Logger,
	proxyServerConfigs config.ProxyServerConfigs,
	requestCache request.Storage,
	ruleService rule.Service,
) (http.Handler, error) {
	servers, err := domain.ConvertProxyServers(proxyServerConfigs)
	if err != nil {
		return nil, err
	}
	proxyService := NewService(logger, servers, ruleService)

	return NewProxyHandler(logger, proxyService, requestCache)
}

type defaultHandler struct {
	logger *zerolog.Logger
}

func (h defaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("received request")

	w.Write([]byte("Hello World\n"))
}
