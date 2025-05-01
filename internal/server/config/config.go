// Package config for server configure.
package config

import (
	"flag"
	"fmt"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/Kopleman/metcol/internal/common/utils"
	"github.com/caarlos0/env/v6"
)

const defaultStoreInterval int64 = 300
const defaultFileStoragePath string = "./store.json"
const defaultRestoreVal bool = true
const defaultCPUProfilePath string = "./profiles/cpuprofile.pprof"
const defaultMemProfilePath string = "./profiles/memprofile.pprof"
const defaultAddress string = "localhost:8080"

// Config contains all settled via envs or flags params.
type Config struct {
	NetAddr             *flags.NetAddress // server address
	FileStoragePath     string            // path to file for mem-store dump
	DataBaseDSN         string            // DSN of postgres DSN
	Key                 string            // hash key for sign received data
	ProfilerCPUFilePath string            // where to store CPU profile
	ProfilerMemFilePath string            // where to store mem profile
	PrivateKeyPath      string            // path to private key
	StoreInterval       int64             // how often dump memo store to file
	ProfilerCollectTime int64             // how long to collect data after start-up
	Restore             bool              // restore memo-store from file
	TrustedSubnet       string            // CIDR for filtering requests
}

type configFromSource struct {
	Restore             *bool  `json:"restore" env:"RESTORE"`
	EndPoint            string `json:"address" env:"ADDRESS"`
	FileStoragePath     string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DataBaseDSN         string `json:"database_dsn" env:"DATABASE_DSN"`
	Key                 string `json:"key" env:"KEY"`
	ProfilerCPUFilePath string `json:"profiler_cpu_file_path" env:"PROFILER_CPU_FILE_PATH"`
	ProfilerMemFilePath string `json:"profiler_mem_file_path" env:"PROFILER_MEM_FILE_PATH"`
	PrivateKeyPath      string `json:"crypto_key" env:"PRIVATE_KEY_PATH"`
	StoreInterval       int64  `json:"store_interval" env:"STORE_INTERVAL"`
	ProfilerCollectTime int64  `json:"profiler_collect_time" env:"PROFILER_COLLECT_TIME"`
	TrustedSubnet       string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
}

func applyConfigFromSource(source *configFromSource, config *Config) error {
	if source.EndPoint != "" {
		if err := config.NetAddr.Set(source.EndPoint); err != nil {
			return fmt.Errorf("failed to set endpoint address for agent from json data: %w", err)
		}
	}

	if source.StoreInterval > 0 {
		config.StoreInterval = source.StoreInterval
	}

	if source.FileStoragePath != "" {
		config.FileStoragePath = source.FileStoragePath
	}

	if source.Restore != nil {
		config.Restore = *source.Restore
	}

	if source.DataBaseDSN != "" {
		config.DataBaseDSN = source.DataBaseDSN
	}

	if source.Key != "" {
		config.Key = source.Key
	}

	if source.ProfilerCollectTime > 0 {
		config.ProfilerCollectTime = source.ProfilerCollectTime
	}

	if source.ProfilerCPUFilePath != "" {
		config.ProfilerCPUFilePath = source.ProfilerCPUFilePath
	}

	if source.ProfilerMemFilePath != "" {
		config.ProfilerMemFilePath = source.ProfilerMemFilePath
	}

	return nil
}

func applyConfigFromFlags(cfgFromFlags *configFromSource, config *Config) error {
	if err := applyConfigFromSource(cfgFromFlags, config); err != nil {
		return fmt.Errorf("failed to apply config from flags: %w", err)
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
		return fmt.Errorf("error applying config from json data: %w", err)
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

// ParseServerConfig produce config for server via parsing env and flags(envs preferred).
func ParseServerConfig() (*Config, error) {
	cfgFromFlags := new(configFromSource)
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = "8080"
	config.NetAddr = netAddr
	config.ProfilerCPUFilePath = defaultCPUProfilePath
	config.ProfilerMemFilePath = defaultMemProfilePath

	flag.StringVar(&cfgFromFlags.EndPoint, "a", defaultAddress, "address and port of collector-server")

	flag.Int64Var(&config.StoreInterval, "i", defaultStoreInterval, "store interval")

	flag.StringVar(&config.FileStoragePath, "f", defaultFileStoragePath, "store file path")

	flag.BoolVar(&config.Restore, "r", defaultRestoreVal, "restore store")

	flag.StringVar(&config.DataBaseDSN, "d", "", "database DSN")

	flag.StringVar(&config.Key, "k", "", "cypher key")

	flag.StringVar(&config.PrivateKeyPath, "crypto-key", "", "cypher key")

	flag.StringVar(&config.ProfilerCPUFilePath, "t", "", "profiler cpu filename")

	pathToConfig := flag.String("c", "", "CIDR for filtering requests")

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
