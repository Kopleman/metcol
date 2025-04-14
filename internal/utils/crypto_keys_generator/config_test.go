package cryptokeysgenerator

import (
	"flag"
	"os"
	"testing"
)

func TestParseConfig(t *testing.T) {
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	tests := []struct {
		name        string
		args        []string
		want        *Config
		wantErr     string
		expectError bool
	}{
		{
			name:    "valid flags",
			args:    []string{"cmd", "-p", "public.pem", "-r", "private.pem"},
			want:    &Config{PublicKeyPath: "public.pem", PrivateKeyPath: "private.pem"},
			wantErr: "",
		},
		{
			name:        "missing public key",
			args:        []string{"cmd", "-r", "private.pem"},
			wantErr:     "public key and private key are required",
			expectError: true,
		},
		{
			name:        "missing private key",
			args:        []string{"cmd", "-p", "public.pem"},
			wantErr:     "public key and private key are required",
			expectError: true,
		},
		{
			name:        "missing both flags",
			args:        []string{"cmd"},
			wantErr:     "public key and private key are required",
			expectError: true,
		},
		{
			name:        "empty values",
			args:        []string{"cmd", "-p", "", "-r", ""},
			wantErr:     "public key and private key are required",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем состояние флагов перед каждым тестом
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ExitOnError)
			os.Args = tt.args

			got, err := ParseConfig()
			if (err != nil) != tt.expectError {
				t.Errorf("ParseConfig() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectError {
				if err.Error() != tt.wantErr {
					t.Errorf("ParseConfig() error = %v, wantErr %v", err.Error(), tt.wantErr)
				}
				return
			}

			if got.PublicKeyPath != tt.want.PublicKeyPath || got.PrivateKeyPath != tt.want.PrivateKeyPath {
				t.Errorf("ParseConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseConfig_FlagError(t *testing.T) {
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Тест на некорректный флаг
	os.Args = []string{"cmd", "-invalid"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Перенаправляем вывод ошибок, чтобы не засорять тесты
	defer func() { flag.CommandLine.SetOutput(nil) }()
	flag.CommandLine.SetOutput(os.NewFile(0, os.DevNull))

	_, err := ParseConfig()
	if err == nil {
		t.Error("Expected error for invalid flag, got nil")
	}
}
