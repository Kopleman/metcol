package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
)

const defaultStoreInterval int64 = 300
const defaultFileStoragePath string = "./store.json"
const defaultRestoreVal bool = true

type Config struct {
	NetAddr         *flags.NetAddress
	FileStoragePath string
	DataBaseDSN     string
	StoreInterval   int64
	Restore         bool
}

type configFromEnv struct {
	Restore         *bool  `env:"RESTORE"`
	EndPoint        string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
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

	flag.Int64Var(&config.StoreInterval, "i", defaultStoreInterval, "store interval")

	flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "store file path")

	flag.BoolVar(&config.Restore, "r", defaultRestoreVal, "restore store")

	flag.StringVar(&config.DataBaseDSN, "d", "", "database DSN")

	flag.Parse()

	if config.StoreInterval < 0 {
		return nil, fmt.Errorf("invalid store interval value prodived via flag: %v", config.StoreInterval)
	}

	if err := env.Parse(cfgFromEnv); err != nil {
		return nil, fmt.Errorf("failed to parse server envs: %w", err)
	}

	if cfgFromEnv.EndPoint != "" {
		if err := netAddr.Set(cfgFromEnv.EndPoint); err != nil {
			return nil, fmt.Errorf("failed to set endpoint address for server: %w", err)
		}
	}

	if cfgFromEnv.StoreInterval < 0 {
		return nil, fmt.Errorf("invalid store interval value prodived via envs: %v", cfgFromEnv.StoreInterval)
	}

	if cfgFromEnv.StoreInterval > 0 {
		config.StoreInterval = cfgFromEnv.StoreInterval
	}

	if cfgFromEnv.FileStoragePath != "" {
		config.FileStoragePath = cfgFromEnv.FileStoragePath
	}

	if cfgFromEnv.Restore != nil {
		config.Restore = *cfgFromEnv.Restore
	}

	if cfgFromEnv.DataBaseDSN != "" {
		config.DataBaseDSN = cfgFromEnv.DataBaseDSN
	}

	return config, nil
}
