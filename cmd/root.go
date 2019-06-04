package cmd

import (
	"fmt"
	"github.com/aldas/xroad-mock-proxy/cmd/version"
	"github.com/aldas/xroad-mock-proxy/pkg/config"
	"github.com/aldas/xroad-mock-proxy/pkg/mock"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/api"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/domain"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/request"
	"github.com/aldas/xroad-mock-proxy/pkg/proxy/rule"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
)

const (
	defaultConfigName string = ".xroad-mock-proxy"
	// EnvPrefix is prefix for environment variables for loaded conf variables
	EnvPrefix string = "XMP"

	// flagConfigFile is flag for configuration file location
	flagConfigFile string = "config"
)

// RootCmd is parent command for all commands
type RootCmd struct {
	*cobra.Command
	logger *zerolog.Logger
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func (c *RootCmd) Execute() {
	if err := c.Command.Execute(); err != nil {
		c.logger.Error().Err(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix(EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

func addSubcommands(cmd *RootCmd, logger *zerolog.Logger, versionStr string, buildStr string) {
	cmd.AddCommand(version.NewVersionCmd(versionStr, buildStr))
}

// NewRootCmd creates new root command
func NewRootCmd(logger *zerolog.Logger, version string, build string) *RootCmd {
	var cmd = &RootCmd{
		&cobra.Command{
			Use:   "xroad-mock-proxy",
			Short: "X-road mock proxy",
			Run: func(cmd *cobra.Command, args []string) {
				err := viper.BindPFlags(cmd.Flags())
				if err != nil {
					logger.Fatal().Err(err).Msg("failed to bind flags")
				}

				appConfig := loadConfig(logger)

				logger.Info().Str("version", version).Str("build", build).Msg("Starting application")
				logger.Debug().Interface("conf", appConfig).Msg("configuration is")

				var wg sync.WaitGroup
				proxyLogger := logger.With().Str("app", "proxy").Logger()
				startProxy(&proxyLogger, appConfig, &wg)

				if appConfig.MockConf.Enabled {
					mockLogger := logger.With().Str("app", "mock").Logger()
					mock.Start(&mockLogger, appConfig.MockConf, &wg)
				}

				wg.Wait()
			},
		},
		logger,
	}

	addSubcommands(cmd, logger, version, build)

	cobra.OnInitialize(initConfig)

	usageStr := fmt.Sprintf("config file (default is $HOME/%v.yaml)", defaultConfigName)
	cmd.PersistentFlags().String(flagConfigFile, "", usageStr)

	return cmd
}

func startProxy(logger *zerolog.Logger, appConfig config.Config, wg *sync.WaitGroup) {
	proxyConf := appConfig.ProxyConf
	routesConf := proxyConf.RoutesConf

	rules, err := domain.ConvertRules(routesConf.Rules)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to bind flags")
	}

	rulesStorageConf := proxyConf.StorageConf.Rules
	ruleService := rule.NewService(logger, rule.NewStorage(logger, rules, rulesStorageConf.Size, rulesStorageConf.Expiration))

	requestStorageConf := proxyConf.StorageConf.Requests
	requestCache := request.NewStorage(requestStorageConf.Size, requestStorageConf.Expiration)
	proxy.Start(logger, proxyConf.ServerConf, routesConf.Servers, requestCache, ruleService, wg)
	if proxyConf.APIConf.Enabled {
		api.Start(logger, proxyConf.APIConf, requestCache, ruleService, wg)
	}
}

func loadConfig(logger *zerolog.Logger) config.Config {
	configFile := viper.GetString(flagConfigFile)
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName(defaultConfigName)
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Fatal().Err(err).Msg("failed to read configuration file")
		}
		if configFile != "" {
			logger.Fatal().Err(err).Str("flagConfigFile", configFile).Msg("failed load config")
		}
	}

	var appConfig config.Config
	err := viper.Unmarshal(&appConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed load app config")
	}

	return appConfig
}
