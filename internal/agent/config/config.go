// Package config for agent configure.
package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/Kopleman/metcol/internal/common/utils"
	"github.com/caarlos0/env/v6"
)

const defaultReportInterval int64 = 10
const defaultPollInterval int64 = 2
const defaultRateInterval int64 = 10
const defaultAddress string = "localhost:8080"

// Config contains all settled via envs or flags params.
type Config struct {
	EndPoint       *flags.NetAddress // where agent will send metrics
	Key            string            // hash key for sign sent data
	GRPCEndPoint   *flags.NetAddress // where agent will send metrics via grpc
	PublicKeyPath  string            // path to public key
	ReportInterval int64             // how often data will be sent
	PollInterval   int64             // how often metrics will be collected
	RateLimit      int64             // limits number of workers for sending
}

type configFromSource struct {
	EndPoint       string `json:"address" env:"ADDRESS"`
	Key            string `json:"key" env:"KEY"`
	GRPCEndPoint   string `json:"grpc_address" env:"GRPC_ADDRESS"`
	PublicKeyPath  string `json:"crypto_key" env:"KEY_PATH"`
	ReportInterval int64  `json:"report_interval" env:"REPORT_INTERVAL"`
	PollInterval   int64  `json:"poll_interval" env:"POLL_INTERVAL"`
	RateLimit      int64  `json:"rate_limit" env:"RATE_LIMIT"`
}

func applyConfigFromSource(source *configFromSource, config *Config) error {
	if source.EndPoint != "" {
		if err := config.EndPoint.Set(source.EndPoint); err != nil {
			return fmt.Errorf("failed to set endpoint address for agent: %w", err)
		}
	}

	if source.GRPCEndPoint != "" {
		if err := config.GRPCEndPoint.Set(source.GRPCEndPoint); err != nil {
			return fmt.Errorf("failed to set grpc endpoint address for agent: %w", err)
		}
	}

	if source.PollInterval < 0 {
		return fmt.Errorf("invalid poll interval value prodived via envs: %v", source.PollInterval)
	}

	if source.ReportInterval < 0 {
		return fmt.Errorf("invalid report interval value prodived via envs: %v", source.ReportInterval)
	}

	if source.PollInterval > 0 {
		config.PollInterval = source.PollInterval
	}

	if source.ReportInterval > 0 {
		config.ReportInterval = source.ReportInterval
	}

	if source.Key != "" {
		config.Key = source.Key
	}

	if source.RateLimit > 0 {
		config.RateLimit = source.RateLimit
	}

	return nil
}

func applyConfigFromFlags(cfgFromFlags *configFromSource, config *Config) error {
	if cfgFromFlags.EndPoint != "" {
		if err := config.EndPoint.Set(cfgFromFlags.EndPoint); err != nil {
			return fmt.Errorf("failed to set endpoint address for agent: %w", err)
		}
	}

	if cfgFromFlags.ReportInterval < 0 {
		return fmt.Errorf("invalid report interval value prodived via flag: %v", cfgFromFlags.ReportInterval)
	}

	if cfgFromFlags.PollInterval < 0 {
		return fmt.Errorf("invalid poll interval value prodived via flag: %v", cfgFromFlags.PollInterval)
	}

	if cfgFromFlags.Key != "" {
		config.Key = cfgFromFlags.Key
	}
	if cfgFromFlags.PublicKeyPath != "" {
		config.PublicKeyPath = cfgFromFlags.PublicKeyPath
	}
	if cfgFromFlags.ReportInterval != 0 {
		config.ReportInterval = cfgFromFlags.ReportInterval
	}
	if cfgFromFlags.PollInterval != 0 {
		config.PollInterval = cfgFromFlags.PollInterval
	}
	if cfgFromFlags.RateLimit != 0 {
		config.RateLimit = cfgFromFlags.RateLimit
	}

	return nil
}

func applyConfigFromJSON(pathToConfigFile string, config *Config) error {
	cfgFromJSON := new(configFromSource)
	if pathToConfigFile == "" {
		return nil
	}

	if err := utils.GetConfigFromFile(pathToConfigFile, cfgFromJSON); err != nil {
		return fmt.Errorf("error reading config from file: %w", err)
	}
	if err := applyConfigFromSource(cfgFromJSON, config); err != nil {
		return fmt.Errorf("error applying config from json-file: %w", err)
	}

	return nil
}

func applyConfigFromEnv(config *Config) error {
	cfgFromEnv := new(configFromSource)
	if err := env.Parse(cfgFromEnv); err != nil {
		return fmt.Errorf("failed to parse agent envs: %w", err)
	}
	if err := applyConfigFromSource(cfgFromEnv, config); err != nil {
		return fmt.Errorf("failed to apply config from env: %w", err)
	}
	return nil
}

// ParseAgentConfig produce config for agent via parsing env and flags(envs preferred).
func ParseAgentConfig() (*Config, error) {
	cfgFromFlags := new(configFromSource)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = "8080"
	config.EndPoint = netAddr

	flag.StringVar(&cfgFromFlags.EndPoint, "a", defaultAddress, "address and port of collector-server")

	flag.Int64Var(&cfgFromFlags.ReportInterval, "r", defaultReportInterval, "report interval")

	flag.Int64Var(&cfgFromFlags.PollInterval, "p", defaultPollInterval, "poll interval")

	flag.StringVar(&cfgFromFlags.Key, "k", "", "cypher key")

	flag.Int64Var(&cfgFromFlags.RateLimit, "l", defaultRateInterval, "output rate interval")

	flag.StringVar(&cfgFromFlags.PublicKeyPath, "crypto-key", "", "cypher key")

	pathToConfig := flag.String("c", "", "Path to config file")

	flag.Parse()

	if err := applyConfigFromJSON(*pathToConfig, config); err != nil {
		return nil, fmt.Errorf("error applaing config from json-file: %w", err)
	}

	if err := applyConfigFromFlags(cfgFromFlags, config); err != nil {
		return nil, fmt.Errorf("error applying config from flags: %w", err)
	}

	if err := applyConfigFromEnv(config); err != nil {
		return nil, fmt.Errorf("error applying config from env: %w", err)
	}

	return config, nil
}
