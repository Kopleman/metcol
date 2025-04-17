package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetConfigFromFile(path string, dest any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close() ///nolint:all // close

	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(dest); decodeErr != nil {
		return fmt.Errorf("could not decode config file: %w", decodeErr)
	}

	return nil
}
