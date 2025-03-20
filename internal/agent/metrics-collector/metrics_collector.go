package metricscollector

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/common/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func (mc *MetricsCollector) GetState() map[string]MetricItem {
	return mc.currentMetricState
}

type CollectResult struct {
	metrics map[string]MetricItem
	err     error
}

func (mc *MetricsCollector) getMemStatMetrics(resultCh chan CollectResult) {
	defer close(resultCh)
	result := CollectResult{}
	result.metrics = make(map[string]MetricItem)
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	result.metrics["Alloc"] = MetricItem{
		value:      strconv.FormatUint(memstats.Alloc, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["BuckHashSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.BuckHashSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["Frees"] = MetricItem{
		value:      strconv.FormatUint(memstats.Frees, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["GCCPUFraction"] = MetricItem{
		value:      strconv.FormatFloat(memstats.GCCPUFraction, 'f', -1, 64),
		metricType: common.GaugeMetricType,
	}
	result.metrics["GCSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.GCSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapAlloc"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapAlloc, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapIdle"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapIdle, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapInuse"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapInuse, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapObjects"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapObjects, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapReleased"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapReleased, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["HeapSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.HeapSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["LastGC"] = MetricItem{
		value:      strconv.FormatUint(memstats.LastGC, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["Lookups"] = MetricItem{
		value:      strconv.FormatUint(memstats.Lookups, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["MCacheInuse"] = MetricItem{
		value:      strconv.FormatUint(memstats.MCacheInuse, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["MCacheSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.MCacheSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["MSpanInuse"] = MetricItem{
		value:      strconv.FormatUint(memstats.MSpanInuse, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["MSpanSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.MSpanSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["Mallocs"] = MetricItem{
		value:      strconv.FormatUint(memstats.Mallocs, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["NextGC"] = MetricItem{
		value:      strconv.FormatUint(memstats.NextGC, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["NumForcedGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(memstats.NumForcedGC), 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["NumGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(memstats.NumGC), 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["OtherSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.OtherSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["PauseTotalNs"] = MetricItem{
		value:      strconv.FormatUint(memstats.PauseTotalNs, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["StackInuse"] = MetricItem{
		value:      strconv.FormatUint(memstats.StackInuse, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["StackSys"] = MetricItem{
		value:      strconv.FormatUint(memstats.StackSys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["Sys"] = MetricItem{
		value:      strconv.FormatUint(memstats.Sys, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["TotalAlloc"] = MetricItem{
		value:      strconv.FormatUint(memstats.TotalAlloc, 10),
		metricType: common.GaugeMetricType,
	}
	resultCh <- result
}

func (mc *MetricsCollector) getGopsutilMetrics(resultCh chan CollectResult) {
	defer close(resultCh)
	result := CollectResult{}

	v, err := mem.VirtualMemory()
	if err != nil {
		result.err = fmt.Errorf("could not get virtual memory info: %w", err)
		resultCh <- result
		return
	}
	usages, errPercent := cpu.Percent(0, false)
	if errPercent != nil {
		result.err = fmt.Errorf("could not get cpu usage: %w", err)
		resultCh <- result
		return
	}
	result.metrics = make(map[string]MetricItem)

	result.metrics["TotalMemory"] = MetricItem{
		value:      strconv.FormatUint(v.Total, 10),
		metricType: common.GaugeMetricType,
	}
	result.metrics["FreeMemory"] = MetricItem{
		value:      strconv.FormatUint(v.Free, 10),
		metricType: common.GaugeMetricType,
	}

	for index, usage := range usages {
		metricName := fmt.Sprintf("CPUutilization%d", index+1)
		result.metrics[metricName] = MetricItem{
			value:      strconv.FormatUint(uint64(usage), 10),
			metricType: common.GaugeMetricType,
		}
	}
	resultCh <- result
}

func (mc *MetricsCollector) CollectAllMetrics() error {
	gopsutilChan := make(chan CollectResult)
	memstatChan := make(chan CollectResult)

	go mc.getMemStatMetrics(memstatChan)
	go mc.getGopsutilMetrics(gopsutilChan)
	for result := range utils.FanIn(memstatChan, gopsutilChan) {
		if result.err != nil {
			fmt.Printf("CollectAllMetrics error: %s\n", result.err)
			return result.err
		}
		mc.mu.Lock()
		maps.Copy(mc.currentMetricState, result.metrics)
		mc.mu.Unlock()
	}

	if err := mc.increasePollCounter(); err != nil {
		return err
	}

	mc.assignNewRandomValue()

	return nil
}

func (mc *MetricsCollector) increasePollCounter() error {
	currentPCValue, err := strconv.ParseInt(mc.currentMetricState[pollCountMetricName].value, 10, 64)
	if err != nil {
		return fmt.Errorf(
			"unable to parse pollcount value ('%s') on poll counter inc",
			mc.currentMetricState[pollCountMetricName].value,
		)
	}

	mc.currentMetricState[pollCountMetricName] = MetricItem{
		value:      strconv.FormatInt(currentPCValue+1, 10),
		metricType: common.CounterMetricType,
	}

	return nil
}

func (mc *MetricsCollector) resetPollCounter() {
	mc.currentMetricState[pollCountMetricName] = MetricItem{
		value:      "0",
		metricType: common.CounterMetricType,
	}
}

func (mc *MetricsCollector) assignNewRandomValue() {
	mc.currentMetricState[randomValueMetricName] = MetricItem{
		value:      strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		metricType: common.GaugeMetricType,
	}
}

type sendMetricResult struct {
	err      error
	workerID int
}

type sendMetricJob struct {
	name   string
	metric MetricItem
}

func (mc *MetricsCollector) sendMetricsViaWorkers() error {
	metricsCount := len(mc.currentMetricState)

	sendJobs := make(chan sendMetricJob, metricsCount)
	results := make(chan sendMetricResult, metricsCount)
	maxWorkerCount := int(mc.cfg.RateLimit)

	for w := 1; w <= maxWorkerCount; w++ {
		go mc.sendMetricWorker(w, sendJobs, results)
	}

	mc.mu.Lock()
	for name, item := range mc.currentMetricState {
		sendJobs <- sendMetricJob{name: name, metric: item}
	}
	mc.mu.Unlock()
	close(sendJobs)

	numOfDoneJobs := 0
	for result := range results {
		numOfDoneJobs++
		if result.err != nil {
			close(results)
			return fmt.Errorf("sendMetricsViaWorkers error: %w", result.err)
		}
		if numOfDoneJobs == metricsCount {
			close(results)
		}
	}

	return nil
}

func (mc *MetricsCollector) sendMetricWorker(workerID int, jobs <-chan sendMetricJob, results chan<- sendMetricResult) {
	for j := range jobs {
		result := sendMetricResult{
			workerID: workerID,
		}
		if err := mc.sendMetricItem(j.name, j.metric); err != nil {
			result.err = fmt.Errorf("send worker: %w", err)
		}
		results <- result
	}
}

func (mc *MetricsCollector) convertMetricItemToDto(name string, item MetricItem) (*dto.MetricDTO, error) {
	metricDto := &dto.MetricDTO{
		ID:    name,
		MType: item.metricType,
	}
	switch item.metricType {
	case common.CounterMetricType:
		parsedDelta, err := strconv.ParseInt(item.value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value ('%s') for metric '%s': %w", item.value, name, err)
		}
		metricDto.Delta = &parsedDelta
	case common.GaugeMetricType:
		parsedValue, err := strconv.ParseFloat(item.value, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value ('%s') for metric '%s': %w", item.value, name, err)
		}
		metricDto.Value = &parsedValue
	default:
		return nil, fmt.Errorf("unknown metric type: %s", item.metricType)
	}

	return metricDto, nil
}

func (mc *MetricsCollector) sendMetricItem(name string, item MetricItem) error {
	metricDto, err := mc.convertMetricItemToDto(name, item)
	if err != nil {
		return err
	}
	body, marshalErr := json.Marshal(metricDto)
	if marshalErr != nil {
		return fmt.Errorf("unable to marshal metric dto: %w", marshalErr)
	}
	url := "/update"
	respBytes, sendErr := mc.client.Post(url, "application/json", body)
	if sendErr != nil {
		return fmt.Errorf("unable to sent %s metric: %w", name, sendErr)
	}

	var t interface{}
	if err = json.Unmarshal(respBytes, &t); err != nil {
		return fmt.Errorf("unable to unmarshal metric response: %w", err)
	}

	return nil
}

// SendMetrics sends all metrics to config.Endpoint.
func (mc *MetricsCollector) SendMetrics() error {
	metricsBatch := make([]*dto.MetricDTO, 0, len(mc.currentMetricState))
	for name, item := range mc.currentMetricState {
		metricDto, err := mc.convertMetricItemToDto(name, item)
		if err != nil {
			return err
		}
		metricsBatch = append(metricsBatch, metricDto)
	}

	if len(metricsBatch) == 0 {
		return nil
	}

	body, marshalErr := json.Marshal(metricsBatch)
	if marshalErr != nil {
		return fmt.Errorf("unable to marshal metrics batch: %w", marshalErr)
	}

	url := "/updates"
	respBytes, sendErr := mc.client.Post(url, "application/json", body)
	if sendErr != nil {
		return fmt.Errorf("unable to sent metrics batch: %w", sendErr)
	}

	var t interface{}
	if err := json.Unmarshal(respBytes, &t); err != nil {
		return fmt.Errorf("unable to unmarshal metric response: %w", err)
	}

	mc.resetPollCounter()

	return nil
}

func (mc *MetricsCollector) genCollectJobParamsChan(tickerChan <-chan time.Time, args *jobsArg) chan struct{} {
	collectIntervalChan := make(chan struct{})

	go func() {
		defer close(collectIntervalChan)
		for currentTickerTime := range tickerChan {
			if currentTickerTime.After(args.nextJobTime) || currentTickerTime.Equal(args.nextJobTime) {
				args.nextJobTime = currentTickerTime.Add(args.interval)
				collectIntervalChan <- struct{}{}
			}
		}
	}()

	return collectIntervalChan
}

func (mc *MetricsCollector) genSendMetricsJobChan(tickerChan <-chan time.Time, args *jobsArg) chan struct{} {
	reportIntervalChan := make(chan struct{})

	go func() {
		defer close(reportIntervalChan)
		for currentTickerTime := range tickerChan {
			if currentTickerTime.After(args.nextJobTime) || currentTickerTime.Equal(args.nextJobTime) {
				args.nextJobTime = currentTickerTime.Add(args.interval)
				reportIntervalChan <- struct{}{}
			}
		}
	}()

	return reportIntervalChan
}

type collectIntervalJobResults struct {
	jobError error
}

type jobsArg struct {
	nextJobTime time.Time
	interval    time.Duration
}

// Handler performs all agent work - collecting and sending data.
func (mc *MetricsCollector) Handler(sig chan os.Signal) error {
	mc.logger.Info("Starting collect metrics")
	pollTicker := time.NewTicker(1 * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(1 * time.Second)
	defer reportTicker.Stop()

	now := time.Now()
	resultChan := make(chan collectIntervalJobResults)
	defer close(resultChan)

	pollDuration := time.Duration(mc.cfg.PollInterval) * time.Second
	reportDuration := time.Duration(mc.cfg.ReportInterval) * time.Second

	collectJobArgs := jobsArg{
		nextJobTime: now.Add(pollDuration),
		interval:    pollDuration,
	}
	reportJobArgs := jobsArg{
		nextJobTime: now.Add(reportDuration),
		interval:    reportDuration,
	}

	collectIntervalChan := mc.genCollectJobParamsChan(pollTicker.C, &collectJobArgs)
	sendIntervalChan := mc.genSendMetricsJobChan(reportTicker.C, &reportJobArgs)

	go mc.collectIntervalJob(collectIntervalChan, resultChan)
	go mc.sendMetricsIntervalJob(sendIntervalChan, resultChan)

	for {
		select {
		case res := <-resultChan:
			if res.jobError != nil {
				return fmt.Errorf("metrics job interval: %w", res.jobError)
			}
		case <-sig:
			return nil
		}
	}
}

func (mc *MetricsCollector) collectIntervalJob(jobArgsCh <-chan struct{}, outputChan chan collectIntervalJobResults) {
	for range jobArgsCh {
		results := collectIntervalJobResults{}
		mc.logger.Info("collecting metrics")
		err := mc.CollectAllMetrics()
		if err != nil {
			results.jobError = fmt.Errorf("collect metrics interval: %w", err)
		}
		outputChan <- results
	}
}

func (mc *MetricsCollector) sendMetricsIntervalJob(
	jobArgsCh <-chan struct{},
	outputChan chan collectIntervalJobResults,
) {
	for range jobArgsCh {
		results := collectIntervalJobResults{}
		mc.logger.Info("sending metrics")
		err := mc.sendMetricsViaWorkers()
		if err != nil {
			results.jobError = fmt.Errorf("send metrics interval: %w", err)
		}
		mc.logger.Info("sending metrics")
		outputChan <- results
	}
}

type HTTPClient interface {
	Post(url, contentType string, bodyBytes []byte) ([]byte, error)
}

type MetricsCollector struct {
	cfg                *config.Config
	currentMetricState map[string]MetricItem
	client             HTTPClient
	logger             log.Logger
	mu                 *sync.Mutex
}

// NewMetricsCollector creates instance of collector.
func NewMetricsCollector(cfg *config.Config, logger log.Logger, client HTTPClient) *MetricsCollector {
	baseState := map[string]MetricItem{
		pollCountMetricName: {
			value:      "0",
			metricType: common.CounterMetricType,
		},
		randomValueMetricName: {
			value:      "0",
			metricType: common.CounterMetricType,
		},
	}
	return &MetricsCollector{currentMetricState: baseState, client: client, cfg: cfg, logger: logger, mu: &sync.Mutex{}}
}
