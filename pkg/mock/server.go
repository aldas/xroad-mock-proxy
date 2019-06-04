package mock

import (
	"github.com/aldas/xroad-mock-proxy/pkg/common/server"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/api"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/config"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/mock"
	"github.com/aldas/xroad-mock-proxy/pkg/mock/rule"
	"github.com/rs/zerolog"
	"sync"
)

// Start starts X-road mock server in goroutine
func Start(
	logger *zerolog.Logger,
	conf config.MockConf,
	wg *sync.WaitGroup,
) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := serve(logger, conf)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to start X-road mock server")
		}
	}()
}

// Start starts the X-road mock service. Method will block until server is shutdown gracefully or until context timeouts
func serve(logger *zerolog.Logger, conf config.MockConf) error {
	e := server.New()

	rootGroup := e.Group(conf.ContextPath)

	rules, err := domain.ConvertRules(conf.Rules)
	if err != nil {
		return err
	}

	storage := rule.NewStorage(logger, rules, conf.Storage.Size)

	mock.RegisterRoutes(mock.NewService(logger, storage), rootGroup)
	api.RegisterRoutes(rule.NewService(logger, storage), rootGroup)

	if conf.WebAssetsDirectory != "" {
		rootGroup.Static("/", conf.WebAssetsDirectory)
	}

	err = server.Start(e, &server.Config{
		Address:             conf.Address,
		ReadTimeoutSeconds:  conf.ReadTimeoutSeconds,
		WriteTimeoutSeconds: conf.WriteTimeoutSeconds,
		Debug:               conf.Debug,
		DebugPath:           conf.DebugPath,
		TLS: server.TLSConf{
			CAFile:              conf.TLS.CAFile,
			CertFile:            conf.TLS.CertFile,
			KeyFile:             conf.TLS.KeyFile,
			KeyPassword:         conf.TLS.KeyPassword,
			ForceClientCertAuth: conf.TLS.ForceClientCertAuth,
		},
	})

	return err
}
