package config

import (
	"flag"
	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	EndPoint       *flags.NetAddress
	ReportInterval int64
	PollInterval   int64
}

type configFromEnv struct {
	EndPoint       string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func ParseAgentConfig() *Config {
	cfgFromEnv := new(configFromEnv)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = 8080
	config.EndPoint = netAddr

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port of collector-server")

	flag.Int64Var(&config.ReportInterval, "r", 10, "report interval")

	flag.Int64Var(&config.PollInterval, "p", 2, "poll interval")

	flag.Parse()

	if err := env.Parse(cfgFromEnv); err != nil {
		panic(err)
	}

	if cfgFromEnv.EndPoint != "" {
		if err := netAddr.Set(cfgFromEnv.EndPoint); err != nil {
			panic(err)
		}
	}

	if cfgFromEnv.PollInterval > 0 {
		config.PollInterval = cfgFromEnv.PollInterval
	}

	if cfgFromEnv.ReportInterval > 0 {
		config.ReportInterval = cfgFromEnv.ReportInterval
	}

	return config
}
