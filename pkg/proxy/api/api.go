package api

import (
	commonServer "github.com/aldas/xroad-mock-proxy/pkg/common/server"
	requestApi "github.com/aldas/xroad-mock-proxy/pkg/proxy/api/request"
	ruleApi "github.com/aldas/xroad-mock-proxy/pkg/proxy/api/rule"
	serverApi "github.com/aldas/xroad-mock-proxy/pkg/proxy/api/server"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/config"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/server"
	"github.com/labstack/echo"
	"github.com/rs/zerolog"
	"sync"
)

// Start start API service in goroutine
func Start(
	logger *zerolog.Logger,
	conf config.APIConf,
	requestCache request.Storage,
	serverService server.Service,
	ruleService rule.Service,
	wg *sync.WaitGroup,
) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := serve(logger, conf, requestCache, serverService, ruleService)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to start API server")
		}
	}()
}

// Start starts the API service. Method will block until server is shutdown gracefully or until context timeouts
func serve(
	logger *zerolog.Logger,
	conf config.APIConf,
	requestCache request.Storage,
	serverService server.Service,
	ruleService rule.Service,
) error {
	e := commonServer.New()

	rootGroup := e.Group(conf.ContextPath)
	if err := addRoutes(logger, rootGroup, conf, requestCache, serverService, ruleService); err != nil {
		return err
	}

	err := commonServer.Start(e, &commonServer.Config{
		Address:             conf.Address,
		ReadTimeoutSeconds:  conf.ReadTimeoutSeconds,
		WriteTimeoutSeconds: conf.WriteTimeoutSeconds,
		Debug:               conf.Debug,
		DebugPath:           conf.DebugPath,
		TLS: commonServer.TLSConf{
			CAFile:              conf.TLS.CAFile,
			CertFile:            conf.TLS.CertFile,
			KeyFile:             conf.TLS.KeyFile,
			KeyPassword:         conf.TLS.KeyPassword,
			ForceClientCertAuth: conf.TLS.ForceClientCertAuth,
		},
	})

	return err
}

func addRoutes(
	logger *zerolog.Logger,
	rootGroup *echo.Group,
	conf config.APIConf,
	requestCache request.Storage,
	serverService server.Service,
	ruleService rule.Service,
) error {

	if conf.AssetsDirectory != "" {
		rootGroup.Static("/", conf.AssetsDirectory)
	}

	api := rootGroup.Group("/api")

	srv := request.NewService(logger, requestCache)
	requestApi.RegisterRoutes(srv, api)

	ruleApi.RegisterRoutes(ruleService, api)
	serverApi.RegisterRoutes(serverService, api)

	return nil
}
