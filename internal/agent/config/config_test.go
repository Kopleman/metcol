//nolint:all // test cases
package config

import (
	"flag"
	"os"
	"testing"

	"github.com/Kopleman/metcol/internal/common/flags"
	"github.com/stretchr/testify/require"
)

func TestParseAgentConfig(t *testing.T) {
	tests := []struct {
		envs        map[string]string
		want        *Config
		name        string
		args        []string
		expectError bool
	}{
		{
			name: "default values",
			args: []string{},
			want: &Config{
				EndPoint:       &flags.NetAddress{Host: "localhost", Port: "8080"},
				ReportInterval: defaultReportInterval,
				PollInterval:   defaultPollInterval,
				RateLimit:      defaultRateInterval,
			},
		},
		{
			name: "flags override defaults",
			args: []string{
				"-a=127.0.0.1:9090",
				"-r=20",
				"-p=5",
				"-k=secret",
				"-l=5",
			},
			want: &Config{
				EndPoint:       &flags.NetAddress{Host: "127.0.0.1", Port: "9090"},
				Key:            "secret",
				ReportInterval: 20,
				PollInterval:   5,
				RateLimit:      5,
			},
		},
		{
			name: "env override flags",
			args: []string{
				"-a=127.0.0.1:9090",
				"-r=10",
				"-p=2",
				"-k=flagkey",
				"-l=5",
			},
			envs: map[string]string{
				"ADDRESS":         "192.168.1.1:8080",
				"REPORT_INTERVAL": "15",
				"POLL_INTERVAL":   "3",
				"KEY":             "envkey",
				"RATE_LIMIT":      "10",
			},
			want: &Config{
				EndPoint:       &flags.NetAddress{Host: "192.168.1.1", Port: "8080"},
				Key:            "envkey",
				ReportInterval: 15,
				PollInterval:   3,
				RateLimit:      10,
			},
		},
		{
			name:        "negative flag values",
			args:        []string{"-r=-5"},
			want:        nil,
			expectError: true,
		},
		{
			name:        "negative env values",
			envs:        map[string]string{"POLL_INTERVAL": "-3"},
			want:        nil,
			expectError: true,
		},
		{
			name:        "invalid address in env",
			envs:        map[string]string{"ADDRESS": "invalid"},
			want:        nil,
			expectError: true,
		},
		{
			name: "partial env override",
			args: []string{"-r=10", "-p=5"},
			envs: map[string]string{"REPORT_INTERVAL": "20"},
			want: &Config{
				EndPoint:       &flags.NetAddress{Host: "localhost", Port: "8080"},
				ReportInterval: 20,
				PollInterval:   5,
				RateLimit:      defaultRateInterval,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment and flags before each test
			defer func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				os.Clearenv()
			}()

			// Set environment variables
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			// Reset args and parse flags
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = append([]string{"cmd"}, tt.args...)

			// Run test
			got, err := ParseAgentConfig()

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want.EndPoint.String(), got.EndPoint.String())
			require.Equal(t, tt.want.Key, got.Key)
			require.Equal(t, tt.want.ReportInterval, got.ReportInterval)
			require.Equal(t, tt.want.PollInterval, got.PollInterval)
			require.Equal(t, tt.want.RateLimit, got.RateLimit)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		envs    map[string]string
		name    string
		wantErr string
		args    []string
	}{
		{
			name:    "negative report interval flag",
			args:    []string{"-r=-1"},
			wantErr: "invalid report interval value prodived via flag: -1",
		},
		{
			name:    "negative poll interval env",
			envs:    map[string]string{"POLL_INTERVAL": "-5"},
			wantErr: "invalid poll interval value prodived via envs: -5",
		},
		{
			name:    "invalid address format",
			envs:    map[string]string{"ADDRESS": "bad:address:123"},
			wantErr: "failed to set endpoint address for agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
				os.Clearenv()
			}()

			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			os.Args = append([]string{"cmd"}, tt.args...)

			_, err := ParseAgentConfig()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
