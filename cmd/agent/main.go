package main

import (
	"fmt"
	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/agent/metrics-collector"
	"github.com/Kopleman/metcol/internal/common/http-client"
	"time"
)

func main() {
	agentConfig := config.ParseAgentConfig()

	endPointURL := `http://` + agentConfig.EndPoint.String() + `/update/`

	httpClient := httpclient.NewHTTPClient(endPointURL)
	collector := metricscollector.NewMetricsCollector(httpClient)
	now := time.Now()

	pollDuration := time.Duration(agentConfig.PollInterval) * time.Second
	reportDuration := time.Duration(agentConfig.ReportInterval) * time.Second

	collectTimer := now.Add(pollDuration)
	reportTimer := now.Add(reportDuration)

	for {
		time.Sleep(1 * time.Second)

		now = time.Now()
		if now.After(collectTimer) {
			collector.CollectMetrics()
			fmt.Println(fmt.Printf("collected metrics at %s", time.Now().UTC()))
			collectTimer = now.Add(pollDuration)
		}
		if now.After(reportTimer) {
			if err := collector.SendMetrics(); err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(fmt.Printf("sent metrics at %s", time.Now().UTC()))
			reportTimer = now.Add(reportDuration)
		}
	}
}
