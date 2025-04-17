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
	// Создаем временный файл с валидным JSON
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:all // tests

	// Записываем тестовые данные
	content := []byte(`{"key": "test", "value": 42}`)
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Тестируем
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
	// Используем несуществующий файл
	err := GetConfigFromFile("non-existent-file.json", &TestConfig{})
	if err == nil {
		t.Error("Expected error but got nil")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got: %v", err)
	}
}

func TestGetConfigFromFile_DecodeError(t *testing.T) {
	// Создаем временный файл с невалидным JSON
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:all // tests

	// Записываем битые данные
	content := []byte(`{"key": "test", "value": "string instead of number"}`)
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Тестируем
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

func TestGetConfigFromFile_CloseError(t *testing.T) {
	// Создаем временный файл и закрываем его перед использованием
	tmpFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	path := tmpFile.Name()
	tmpFile.Close() //nolint:all // tests
	os.Remove(path) //nolint:all // tests

	// Создаем новый файл с тем же именем, но без прав на чтение
	if err := os.WriteFile(path, []byte("{}"), 0222); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path) //nolint:all // tests

	// Тестируем
	err = GetConfigFromFile(path, &TestConfig{})
	if err == nil {
		t.Error("Expected error but got nil")
	}
}
