package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

const defaultStoreInterval int64 = 300
const defaultFileStoragePath string = "./store.json"
const defaultRestoreVal bool = true
const defaultCPUProfilePath string = "./profiles/cpuprofile.pprof"
const defaultMemProfilePath string = "./profiles/memprofile.pprof"

type Config struct {
	NetAddr             *flags.NetAddress
	FileStoragePath     string
	DataBaseDSN         string
	Key                 string
	ProfilerCPUFilePath string
	ProfilerMemFilePath string
	StoreInterval       int64
	ProfilerCollectTime int64
	Restore             bool
}

type configFromEnv struct {
	Restore             *bool  `env:"RESTORE"`
	EndPoint            string `env:"ADDRESS"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	DataBaseDSN         string `env:"DATABASE_DSN"`
	Key                 string `env:"KEY"`
	ProfilerCPUFilePath string `env:"PROFILER_CPU_FILE_PATH"`
	ProfilerMemFilePath string `env:"PROFILER_MEM_FILE_PATH"`
	StoreInterval       int64  `env:"STORE_INTERVAL"`
	ProfilerCollectTime int64  `env:"PROFILER_COLLECT_TIME"`
}

func ParseServerConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfgFromEnv := new(configFromEnv)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = "8080"
	config.NetAddr = netAddr
	config.ProfilerCPUFilePath = defaultCPUProfilePath
	config.ProfilerMemFilePath = defaultMemProfilePath

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port to run server")

	flag.Int64Var(&config.StoreInterval, "i", defaultStoreInterval, "store interval")

	flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "store file path")

	flag.BoolVar(&config.Restore, "r", defaultRestoreVal, "restore store")

	flag.StringVar(&config.DataBaseDSN, "d", "", "database DSN")

	flag.StringVar(&config.Key, "k", "", "cypher key")

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

	if cfgFromEnv.Key != "" {
		config.Key = cfgFromEnv.Key
	}

	if cfgFromEnv.ProfilerCollectTime > 0 {
		config.ProfilerCollectTime = cfgFromEnv.ProfilerCollectTime
	}

	if cfgFromEnv.ProfilerCPUFilePath != "" {
		config.ProfilerCPUFilePath = cfgFromEnv.ProfilerCPUFilePath
	}

	if cfgFromEnv.ProfilerMemFilePath != "" {
		config.ProfilerMemFilePath = cfgFromEnv.ProfilerMemFilePath
	}

	return config, nil
}
