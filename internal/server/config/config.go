package config

import (
	"flag"
	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	NetAddr *flags.NetAddress
}

type configFromEnv struct {
	EndPoint string `env:"ADDRESS"`
}

func ParseServerConfig() *Config {
	cfgFromEnv := new(configFromEnv)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = 8080
	config.NetAddr = netAddr

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port to run server")
	flag.Parse()

	if err := env.Parse(cfgFromEnv); err != nil {
		panic(err)
	}

	if cfgFromEnv.EndPoint != "" {
		if err := netAddr.Set(cfgFromEnv.EndPoint); err != nil {
			panic(err)
		}
	}

	return config
}
