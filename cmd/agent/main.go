package main

import (
	"fmt"
	httpclient "github.com/Kopleman/metcol/internal/http-client"
	metricscollector "github.com/Kopleman/metcol/internal/metrics-collector"
	"time"
)

func main() {
	httpClient := httpclient.NewHttpClient("http://localhost:8080/update/")
	collector := metricscollector.NewMetricsCollector(httpClient)
	now := time.Now()

	collectorTimer := now.Add(2 * time.Second)
	senderTime := now.Add(10 * time.Second)

	for {
		time.Sleep(1 * time.Second)

		now = time.Now()
		if now.After(collectorTimer) {
			fmt.Println("collector")
			collector.CollectMetrics()
			fmt.Println(fmt.Printf("collected metrics at %s", time.Now().UTC()))
			collectorTimer = now.Add(2 * time.Second)
		}
		if now.After(senderTime) {
			if err := collector.SendMetrics(); err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(fmt.Printf("sent metrics at %s", time.Now().UTC()))
			senderTime = now.Add(10 * time.Second)
		}
	}
}
