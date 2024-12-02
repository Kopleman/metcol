package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	NetAddr *flags.NetAddress
}

type configFromEnv struct {
	EndPoint string `env:"ADDRESS"`
}

func ParseServerConfig() (*Config, error) {
	cfgFromEnv := new(configFromEnv)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = "8080"
	config.NetAddr = netAddr

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port to run server")
	flag.Parse()

	if err := env.Parse(cfgFromEnv); err != nil {
		return nil, fmt.Errorf("failed to parse server envs: %v", err)
	}

	if cfgFromEnv.EndPoint != "" {
		if err := netAddr.Set(cfgFromEnv.EndPoint); err != nil {
			return nil, fmt.Errorf("failed to set endpoint address for server: %v", err)
		}
	}

	return config, nil
}
