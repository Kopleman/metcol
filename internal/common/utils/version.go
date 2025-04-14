package utils

import "fmt"

func PrintVersion(buildVersion, buildDate, buildCommit string) {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}

	fmt.Printf("Build version=%s\n", buildVersion)
	fmt.Printf("Build date=%s\n", buildDate)
	fmt.Printf("Build commit=%s\n", buildCommit)
}
