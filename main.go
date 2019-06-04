package main

import (
	"github.com/aldas/xroad-mock-proxy/cmd"
	"github.com/rs/zerolog"
	"os"
	"time"
)

var (
	version = "undefined"
	build   = "undefined"
)

func main() {
	logger := newLogger()
	cmd.NewRootCmd(logger, version, build).Execute()
}

func newLogger() *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Sample(&zerolog.BurstSampler{
		Burst:       5,
		Period:      500 * time.Millisecond,
		NextSampler: &zerolog.BasicSampler{N: 100},
	})

	return &logger
}
