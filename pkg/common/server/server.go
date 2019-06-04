package server

import (
	"context"
	// expose expvar debug endpoints
	_ "expvar"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	// expose pprof profiling endpoints
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

// Config represents server specific config
type Config struct {
	Address             string
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	Debug               bool
	DebugPath           string
	TLS                 TLSConf
}

// TLSConf is server TLS configuration
type TLSConf struct {
	CAFile              string
	CertFile            string
	KeyFile             string
	KeyPassword         string
	ForceClientCertAuth bool
}

// New instantiates new Echo server
func New() *echo.Echo {
	e := echo.New()

	e.Use(
		middleware.Logger(),
		middleware.Recover(),
		middleware.BodyLimit("2M"),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		}),
	)

	e.GET("/health", healthCheck)

	custErr := &customErrHandler{e: e}
	e.HTTPErrorHandler = custErr.handler

	return e
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

// Start starts echo server and block until server is shutdown
func Start(e *echo.Echo, cfg *Config) error {
	server := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Second * 60,
	}

	err := configureTLSConfig(server, cfg.TLS)
	if err != nil {
		return err
	}

	e.Debug = cfg.Debug

	if cfg.DebugPath != "" {
		e.Logger.Info(fmt.Sprintf("Exposing debug endpoints for expvar/pprof on path: %v", cfg.DebugPath))
		e.GET(cfg.DebugPath+"*", echo.WrapHandler(http.DefaultServeMux))
	}

	// Start server
	go func() {
		if err := e.StartServer(server); err != nil {
			e.Logger.Info("Shutting down the server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	return nil
}
