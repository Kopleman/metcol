package cryptokeysgenerator

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Error("NewGenerator() returned nil")
	}
}

func TestGenerateKeys_EmptyPaths(t *testing.T) {
	g := NewGenerator()
	tests := []struct {
		name        string
		privatePath string
		publicPath  string
		wantErr     string
	}{
		{
			name:        "empty private path",
			privatePath: "",
			publicPath:  "public.pem",
			wantErr:     "publicKeyPath or privateKeyPath is empty",
		},
		{
			name:        "empty public path",
			privatePath: "private.pem",
			publicPath:  "",
			wantErr:     "publicKeyPath or privateKeyPath is empty",
		},
		{
			name:        "both paths empty",
			privatePath: "",
			publicPath:  "",
			wantErr:     "publicKeyPath or privateKeyPath is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.GenerateKeys(tt.privatePath, tt.publicPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestGenerateKeys_SamePaths(t *testing.T) {
	g := NewGenerator()
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "same.pem")

	err := g.GenerateKeys(path, path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	wantErr := "publicKeyPath is equal to privateKeyPath"
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("expected error %q, got %q", wantErr, err.Error())
	}
}

func TestGenerateKeys_Success(t *testing.T) {
	tempDir := t.TempDir()
	privatePath := filepath.Join(tempDir, "private.pem")
	publicPath := filepath.Join(tempDir, "public.pem")

	g := NewGenerator()
	if err := g.GenerateKeys(privatePath, publicPath); err != nil {
		t.Fatalf("GenerateKeys failed: %v", err)
	}

	// Проверка существования файлов
	for _, path := range []string{privatePath, publicPath} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %q was not created", path)
		}
	}

	// Проверка формата ключей
	checkPEMFile(t, privatePath, "RSA PRIVATE KEY")
	checkPEMFile(t, publicPath, "PUBLIC KEY")

	// Проверка валидности ключей
	checkValidRSAPair(t, privatePath, publicPath)
}

func checkPEMFile(t *testing.T, path, expectedType string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatalf("no PEM data found in %s", path)
	}

	if block.Type != expectedType {
		t.Fatalf("unexpected PEM type in %s: got %s, want %s", path, block.Type, expectedType)
	}
}

func checkValidRSAPair(t *testing.T, privatePath, publicPath string) {
	t.Helper()
	privData, err := os.ReadFile(privatePath)
	if err != nil {
		t.Fatal(err)
	}
	privBlock, _ := pem.Decode(privData)
	privKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// Чтение публичного ключа
	pubData, err := os.ReadFile(publicPath)
	if err != nil {
		t.Fatal(err)
	}
	pubBlock, _ := pem.Decode(pubData)
	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	// Проверка соответствия ключей
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("public key is not RSA")
	}

	if !privKey.PublicKey.Equal(rsaPubKey) {
		t.Error("public key does not match private key")
	}
}

func TestGenerateKeys_WriteErrors(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name        string
		privatePath string
		publicPath  string
		wantErr     string
	}{
		{
			name:        "invalid private path",
			privatePath: filepath.Join(tempDir, "nonexistent", "private.pem"),
			publicPath:  filepath.Join(tempDir, "public.pem"),
			wantErr:     "failed to export private key to file",
		},
		{
			name:        "invalid public path",
			privatePath: filepath.Join(tempDir, "private.pem"),
			publicPath:  filepath.Join(tempDir, "nonexistent", "public.pem"),
			wantErr:     "failed to export public key to file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator()
			err := g.GenerateKeys(tt.privatePath, tt.publicPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestGetFileDescriptor_Error(t *testing.T) {
	tempDir := t.TempDir()
	readOnlyPath := filepath.Join(tempDir, "readonly.pem")

	// Создаем файл только для чтения
	if err := os.WriteFile(readOnlyPath, []byte("test"), 0o222); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(readOnlyPath) //nolint:all // tests
	if errs := os.Chmod(readOnlyPath, 0o222); errs != nil {
		t.Fatal(errs)
	}

	g := NewGenerator()
	_, err := g.getFileDescriptor(readOnlyPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	wantErr := "permission denied"
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("expected error containing %q, got %q", wantErr, err.Error())
	}
}
