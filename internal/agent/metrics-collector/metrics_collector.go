package metricscollector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Kopleman/metcol/internal/agent/config"
	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
)

func (mc *MetricsCollector) GetState() map[string]MetricItem {
	return mc.currentMetricState
}

func (mc *MetricsCollector) CollectMetrics() error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mc.currentMetricState["Alloc"] = MetricItem{
		value:      strconv.FormatUint(mem.Alloc, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["BuckHashSys"] = MetricItem{
		value:      strconv.FormatUint(mem.BuckHashSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["Frees"] = MetricItem{
		value:      strconv.FormatUint(mem.Frees, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["GCCPUFraction"] = MetricItem{
		value:      strconv.FormatFloat(mem.GCCPUFraction, 'f', -1, 64),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["GCSys"] = MetricItem{
		value:      strconv.FormatUint(mem.GCSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapAlloc"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapAlloc, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapIdle"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapIdle, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapInuse, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapObjects"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapObjects, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapReleased"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapReleased, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["HeapSys"] = MetricItem{
		value:      strconv.FormatUint(mem.HeapSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["LastGC"] = MetricItem{
		value:      strconv.FormatUint(mem.LastGC, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["Lookups"] = MetricItem{
		value:      strconv.FormatUint(mem.Lookups, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["MCacheInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.MCacheInuse, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["MCacheSys"] = MetricItem{
		value:      strconv.FormatUint(mem.MCacheSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["MSpanInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.MSpanInuse, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["MSpanSys"] = MetricItem{
		value:      strconv.FormatUint(mem.MSpanSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["Mallocs"] = MetricItem{
		value:      strconv.FormatUint(mem.Mallocs, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["NextGC"] = MetricItem{
		value:      strconv.FormatUint(mem.NextGC, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["NumForcedGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(mem.NumForcedGC), 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["NumGC"] = MetricItem{
		value:      strconv.FormatUint(uint64(mem.NumGC), 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["OtherSys"] = MetricItem{
		value:      strconv.FormatUint(mem.OtherSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["PauseTotalNs"] = MetricItem{
		value:      strconv.FormatUint(mem.PauseTotalNs, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["StackInuse"] = MetricItem{
		value:      strconv.FormatUint(mem.StackInuse, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["StackSys"] = MetricItem{
		value:      strconv.FormatUint(mem.StackSys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["Sys"] = MetricItem{
		value:      strconv.FormatUint(mem.Sys, 10),
		metricType: common.GaugeMetricType,
	}
	mc.currentMetricState["TotalAlloc"] = MetricItem{
		value:      strconv.FormatUint(mem.TotalAlloc, 10),
		metricType: common.GaugeMetricType,
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

func (mc *MetricsCollector) SendMetricsByOne() error {
	for name, item := range mc.currentMetricState {
		if err := mc.sendMetricItem(name, item); err != nil {
			return err
		}
	}

	mc.resetPollCounter()

	return nil
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
	respBytes, sendErr := mc.client.Post(url, "application/json", bytes.NewBuffer(body))
	if sendErr != nil {
		return fmt.Errorf("unable to sent %s metric: %w", name, sendErr)
	}

	var t interface{}
	if err = json.Unmarshal(respBytes, &t); err != nil {
		return fmt.Errorf("unable to unmarshal metric response: %w", err)
	}

	return nil
}

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
	respBytes, sendErr := mc.client.Post(url, "application/json", bytes.NewBuffer(body))
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

func (mc *MetricsCollector) Run(sig chan os.Signal) error {
	now := time.Now()

	pollDuration := time.Duration(mc.cfg.PollInterval) * time.Second
	reportDuration := time.Duration(mc.cfg.ReportInterval) * time.Second

	args := intervalJobsArg{
		collectTimer:     now.Add(pollDuration),
		reportTimer:      now.Add(reportDuration),
		pollInterval:     pollDuration,
		reportInterval:   reportDuration,
		reportInProgress: false,
	}

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := mc.doIntervalJobs(&args); err != nil {
				ticker.Stop()
				return err
			}
		case <-sig:
			ticker.Stop()
			return nil
		}
	}
}

type intervalJobsArg struct {
	collectTimer     time.Time
	reportTimer      time.Time
	pollInterval     time.Duration
	reportInterval   time.Duration
	reportInProgress bool
}

func (mc *MetricsCollector) doIntervalJobs(args *intervalJobsArg) error {
	now := time.Now()
	if now.After(args.collectTimer) {
		err := mc.CollectMetrics()

		if err != nil {
			return fmt.Errorf("collect metrics: %w", err)
		}

		mc.logger.Info("collected metrics")

		args.collectTimer = now.Add(args.pollInterval)
	}

	if args.reportInProgress {
		return nil
	}

	if now.After(args.reportTimer) {
		args.reportInProgress = true

		err := mc.SendMetrics()

		if err != nil {
			mc.logger.Error(err)
		} else {
			mc.logger.Info("sent metrics")
		}

		args.reportTimer = now.Add(args.reportInterval)
		args.reportInProgress = false
	}

	return nil
}

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) ([]byte, error)
}

type MetricsCollector struct {
	cfg                *config.Config
	currentMetricState map[string]MetricItem
	client             HTTPClient
	logger             log.Logger
}

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
	return &MetricsCollector{currentMetricState: baseState, client: client, cfg: cfg, logger: logger}
}
