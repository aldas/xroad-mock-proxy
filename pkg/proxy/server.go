package proxy

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/server"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	proxyserver "github.com/aldas/xroad-mock-proxy/pkg/proxy/server"
	"github.com/labstack/echo"
	"github.com/rs/zerolog"
	"net/http"
	"sync"
)

//Start starts proxy server in goroutine
func Start(
	logger *zerolog.Logger,
	serverConfig config.ServerConf,
	requestCache request.Storage,
	serverService proxyserver.AccessorService,
	ruleService rule.Service,
	wg *sync.WaitGroup,
) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := serve(logger, serverConfig, requestCache, serverService, ruleService)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to start proxy server")
		}
	}()
}

// Serve starts http(s) server. Method will block until server is shutdown gracefully or until context timeouts
func serve(
	logger *zerolog.Logger,
	serverConfig config.ServerConf,
	requestCache request.Storage,
	serverService proxyserver.AccessorService,
	ruleService rule.Service,
) error {
	e := server.New()

	e.GET("/", echo.WrapHandler(defaultHandler{logger: logger}))

	proxyHandler, err := createProxyHandler(logger, requestCache, serverService, ruleService)
	if err != nil {
		return err
	}

	contextPath := serverConfig.ContextPath
	if contextPath == "" {
		contextPath = XroadDefaulURL
	}
	e.POST(contextPath, echo.WrapHandler(proxyHandler))

	logger.Info().Msg("start serving proxy server")
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
	requestCache request.Storage,
	serverService proxyserver.AccessorService,
	ruleService rule.Service,
) (http.Handler, error) {
	return NewProxyHandler(logger, serverService, ruleService, requestCache)
}

type defaultHandler struct {
	logger *zerolog.Logger
}

func (h defaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("received request")

	w.Write([]byte("Hello World\n"))
}
