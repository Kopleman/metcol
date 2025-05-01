package cryptokeysgenerator

import (
	"errors"
	"flag"
)

// Config contains all settled via flags params.
type Config struct {
	PublicKeyPath  string
	PrivateKeyPath string
}

// ParseConfig produce config for generator via parsing flags.
func ParseConfig() (*Config, error) {
	config := new(Config)

	flag.StringVar(&config.PublicKeyPath, "p", "", "path to public key")
	flag.StringVar(&config.PrivateKeyPath, "r", "", "path to private key")
	flag.Parse()

	if config.PublicKeyPath == "" || config.PrivateKeyPath == "" {
		return nil, errors.New("public key and private key are required")
	}

	return config, nil
}
