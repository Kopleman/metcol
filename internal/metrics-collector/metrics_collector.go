package metrics_collector

import (
	"bytes"
	htttp_client "github.com/Kopleman/metcol/internal/http-client"
	metcolMetrics "github.com/Kopleman/metcol/internal/metrics"
	"math/rand/v2"
	"runtime"
	"strconv"
)

type IMetricsCollector interface {
	CollectMetrics()
	SendMetrics() error
	GetState() map[string]MetricItem
}

type MetricsCollector struct {
	currentMetricState map[string]MetricItem
	client             htttp_client.IHttpClient
}

func NewMetricsCollector(client htttp_client.IHttpClient) IMetricsCollector {
	baseState := map[string]MetricItem{
		"PollCount": {
			value:      "0",
			metricType: metcolMetrics.CounterMetricType,
		},
		"RandomValue": {
			value:      "0",
			metricType: metcolMetrics.CounterMetricType,
		},
	}
	return &MetricsCollector{currentMetricState: baseState, client: client}
}

func (mc *MetricsCollector) GetState() map[string]MetricItem {
	return mc.currentMetricState
}

func (mc *MetricsCollector) CollectMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mc.currentMetricState["Alloc"] = MetricItem{
		value:      strconv.FormatUint(mem.Alloc, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["BuckHashSys"] = MetricItem{
		value:      strconv.FormatUint(mem.BuckHashSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["Frees"] = MetricItem{
		value:      strconv.FormatUint(mem.Frees, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["GCCPUFraction"] = MetricItem{
		value:      strconv.FormatFloat(mem.GCCPUFraction, 'f', -1, 64),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["GCSys"] = MetricItem{
		value:      strconv.FormatUint(mem.GCSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapAlloc"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapAlloc, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapIdle"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapIdle, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapInuse, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapObjects"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapObjects, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapReleased"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapReleased, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["HeapSys"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["LastGC"] = MetricItem{
		value:      strconv.FormatUint(mem.LastGC, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["Lookups"] = MetricItem{
		value:      strconv.FormatUint(mem.Lookups, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["MCacheInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.MCacheInuse, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["MCacheSys"] = MetricItem{
		value:      strconv.FormatUint(mem.MCacheSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["MSpanInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.MSpanInuse, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["MSpanSys"] = MetricItem{
		value:      strconv.FormatUint(mem.MSpanSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["Mallocs"] = MetricItem{
		value:      strconv.FormatUint(mem.Mallocs, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["NextGC"] = MetricItem{
		value:      strconv.FormatUint(mem.NextGC, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["NumForcedGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(mem.NumForcedGC), 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["NumGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(mem.NumGC), 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["OtherSys"] = MetricItem{
		value:      strconv.FormatUint(mem.OtherSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["PauseTotalNs"] = MetricItem{
		value:      strconv.FormatUint(mem.PauseTotalNs, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["StackInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.StackInuse, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["StackSys"] = MetricItem{
		value:      strconv.FormatUint(mem.StackSys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["Sys"] = MetricItem{
		value:      strconv.FormatUint(mem.Sys, 10),
		metricType: metcolMetrics.GougeMetricType,
	}
	mc.currentMetricState["TotalAlloc"] = MetricItem{
		value:      strconv.FormatUint(mem.TotalAlloc, 10),
		metricType: metcolMetrics.GougeMetricType,
	}

	if err := mc.increasePollCounter(); err != nil {
		panic(err.Error())
	}

	mc.assignNewRandomValue()
}

func (mc *MetricsCollector) increasePollCounter() error {
	currentPCValue, err := strconv.ParseInt(mc.currentMetricState["PollCount"].value, 10, 64)
	if err != nil {
		return ErrCounterPollParse
	}

	mc.currentMetricState["PollCount"] = MetricItem{
		value:      strconv.FormatInt(currentPCValue+1, 10),
		metricType: metcolMetrics.CounterMetricType,
	}

	return nil
}

func (mc *MetricsCollector) assignNewRandomValue() {
	mc.currentMetricState["RandomValue"] = MetricItem{
		value:      strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		metricType: metcolMetrics.GougeMetricType,
	}
}

func (mc *MetricsCollector) SendMetrics() error {
	for name, item := range mc.currentMetricState {
		if err := mc.sendMetricItem(name, item); err != nil {
			return err
		}
	}

	return nil
}

func (mc *MetricsCollector) sendMetricItem(name string, item MetricItem) error {
	url := string(item.metricType) + "/" + name + "/" + item.value

	body := []byte("")
	_, err := mc.client.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return nil
}
