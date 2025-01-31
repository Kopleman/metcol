package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
)

const defaultReportInterval int64 = 10
const defaultPollInterval int64 = 2
const defaultRateInterval int64 = 10

type Config struct {
	EndPoint       *flags.NetAddress
	Key            string
	ReportInterval int64
	PollInterval   int64
	RateLimit      int64
}

type configFromEnv struct {
	EndPoint       string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	RateLimit      int64  `env:"RATE_LIMIT"`
}

func ParseAgentConfig() (*Config, error) {
	cfgFromEnv := new(configFromEnv)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = "8080"
	config.EndPoint = netAddr

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port of collector-server")

	flag.Int64Var(&config.ReportInterval, "r", defaultReportInterval, "report interval")

	flag.Int64Var(&config.PollInterval, "p", defaultPollInterval, "poll interval")

	flag.StringVar(&config.Key, "k", "", "cypher key")

	flag.Int64Var(&config.RateLimit, "l", defaultRateInterval, "output rate interval")

	flag.Parse()

	if config.ReportInterval < 0 {
		return nil, fmt.Errorf("invalid report interval value prodived via flag: %v", config.ReportInterval)
	}

	if config.PollInterval < 0 {
		return nil, fmt.Errorf("invalid poll interval value prodived via flag: %v", config.PollInterval)
	}

	if err := env.Parse(cfgFromEnv); err != nil {
		return nil, fmt.Errorf("failed to parse agent envs: %w", err)
	}

	if cfgFromEnv.EndPoint != "" {
		if err := netAddr.Set(cfgFromEnv.EndPoint); err != nil {
			return nil, fmt.Errorf("failed to set endpoint address for agent: %w", err)
		}
	}

	if cfgFromEnv.PollInterval < 0 {
		return nil, fmt.Errorf("invalid poll interval value prodived via envs: %v", cfgFromEnv.PollInterval)
	}

	if cfgFromEnv.ReportInterval < 0 {
		return nil, fmt.Errorf("invalid report interval value prodived via envs: %v", cfgFromEnv.ReportInterval)
	}

	if cfgFromEnv.PollInterval > 0 {
		config.PollInterval = cfgFromEnv.PollInterval
	}

	if cfgFromEnv.ReportInterval > 0 {
		config.ReportInterval = cfgFromEnv.ReportInterval
	}

	if cfgFromEnv.Key != "" {
		config.Key = cfgFromEnv.Key
	}

	if cfgFromEnv.RateLimit > 0 {
		config.RateLimit = cfgFromEnv.RateLimit
	}

	return config, nil
}
