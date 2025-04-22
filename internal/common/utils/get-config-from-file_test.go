package utils

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type TestConfig struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func TestGetConfigFromFile_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:all // tests

	content := []byte(`{"key": "test", "value": 42}`)
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close() //nolint:all // tests

	var config TestConfig
	err = GetConfigFromFile(tmpFile.Name(), &config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if config.Key != "test" || config.Value != 42 {
		t.Errorf("Config not parsed correctly. Got %+v", config)
	}
}

func TestGetConfigFromFile_OpenError(t *testing.T) {
	err := GetConfigFromFile("non-existent-file.json", &TestConfig{})
	if err == nil {
		t.Error("Expected error but got nil")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got: %v", err)
	}
}

func TestGetConfigFromFile_DecodeError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:all // tests

	content := []byte(`{"key": "test", "value": "string instead of number"}`)
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close() //nolint:all // tests

	var config TestConfig
	err = GetConfigFromFile(tmpFile.Name(), &config)
	if err == nil {
		t.Error("Expected decoding error but got nil")
	}

	var jsonErr *json.UnmarshalTypeError
	if !errors.As(err, &jsonErr) {
		t.Errorf("Expected json.UnmarshalTypeError, got: %T", err)
	}
}

func TestGetConfigFromFile_PermissionError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	path := tmpFile.Name()
	tmpFile.Close() //nolint:all // tests
	os.Remove(path) //nolint:all // tests

	if err = os.WriteFile(path, []byte("{}"), 0o222); err != nil { //nolint:all // tests
		t.Fatal(err)
	}
	if errs := os.Chmod(path, 0o222); errs != nil { //nolint:all // tests
		t.Fatal(errs)
	}
	defer os.Remove(path) //nolint:all // tests

	err = GetConfigFromFile(path, &TestConfig{})
	if err == nil {
		t.Error("Expected error but got nil")
	}
}
