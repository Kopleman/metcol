package config

import (
	"flag"
	"github.com/Kopleman/metcol/internal/common/flags"
)

type Config struct {
	NetAddr *flags.NetAddress
}

func ParseServerConfig() *Config {
	config := new(Config)
	netAddr := new(flags.NetAddress)
	netAddr.Host = "localhost"
	netAddr.Port = 8080
	config.NetAddr = netAddr

	netAddrValue := flag.Value(netAddr)
	flag.Var(netAddrValue, "a", "address and port to run server")
	flag.Parse()

	return config
}
