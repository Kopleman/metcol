// Package generatekeys to generate keys
package main

import (
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
	cryptokeysgenerator "github.com/Kopleman/metcol/internal/utils/crypto_keys_generator"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	utils.PrintVersion(buildVersion, buildDate, buildCommit)

	logger := log.New(
		log.WithAppVersion("local"),
	)

	logger.Info("Starting public/private key generation")
	config, err := cryptokeysgenerator.ParseConfig()
	if err != nil {
		logger.Fatalf("unable to parse config for generator: %w", err)
	}
	generator := cryptokeysgenerator.NewGenerator()
	if genErr := generator.GenerateKeys(config.PrivateKeyPath, config.PublicKeyPath); genErr != nil {
		logger.Fatalf("unable to generate keys: %w", genErr)
	}

	logger.Infof("Keys generated successfully, private=%s, public=%s", config.PrivateKeyPath, config.PublicKeyPath)
}
