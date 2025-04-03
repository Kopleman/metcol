package utils

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w //nolint:reassign // test purpose

	f()

	w.Close()       //nolint:all // test-cases
	os.Stdout = old //nolint:reassign // test purpose

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:all // test-cases

	return buf.String()
}

func TestPrintVersion(t *testing.T) {
	tests := []struct {
		name           string
		buildVersion   string
		buildDate      string
		buildCommit    string
		expectedOutput string
	}{
		{
			name:           "all non-empty",
			buildVersion:   "1.0.0",
			buildDate:      "2023-10-01",
			buildCommit:    "abcd1234",
			expectedOutput: "Build version=1.0.0\nBuild date=2023-10-01\nBuild commit=abcd1234\n",
		},
		{
			name:           "version empty",
			buildVersion:   "",
			buildDate:      "2023-10-01",
			buildCommit:    "abcd1234",
			expectedOutput: "Build version=N/A\nBuild date=2023-10-01\nBuild commit=abcd1234\n",
		},
		{
			name:           "date empty",
			buildVersion:   "1.0.0",
			buildDate:      "",
			buildCommit:    "abcd1234",
			expectedOutput: "Build version=1.0.0\nBuild date=N/A\nBuild commit=abcd1234\n",
		},
		{
			name:           "commit empty",
			buildVersion:   "1.0.0",
			buildDate:      "2023-10-01",
			buildCommit:    "",
			expectedOutput: "Build version=1.0.0\nBuild date=2023-10-01\nBuild commit=N/A\n",
		},
		{
			name:           "all empty",
			buildVersion:   "",
			buildDate:      "",
			buildCommit:    "",
			expectedOutput: "Build version=N/A\nBuild date=N/A\nBuild commit=N/A\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintVersion(tt.buildVersion, tt.buildDate, tt.buildCommit)
			})

			if output != tt.expectedOutput {
				t.Errorf(
					"Test case %s failed.\nExpected:\n%s\nGot:\n%s",
					tt.name,
					tt.expectedOutput,
					output,
				)
			}
		})
	}
}
