package config

import (
	"flag"
	"github.com/Kopleman/metcol/internal/common/flags"
)

type Config struct {
	EndPoint       *flags.NetAddress
	ReportInterval int64
	PollInterval   int64
}

func ParseAgentConfig() *Config {
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

	return config
}
