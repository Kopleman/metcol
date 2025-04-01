// Package dto which contains DTO.
package dto

import (
	"encoding/json"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
)

// GetValueRequest dto for fetching metric data.
type GetValueRequest struct {
	ID    string            `json:"id"`   // metric name.
	MType common.MetricType `json:"type"` // metric type - gauge or counter.
}

func parseType(metricTypeAsString string) (common.MetricType, error) {
	mType, ok := common.StringToMetricType(metricTypeAsString)
	if !ok {
		return common.UnknownMetricType, fmt.Errorf(common.ErrUnknownMetric, metricTypeAsString)
	}

	if mType == common.UnknownMetricType {
		return common.UnknownMetricType, fmt.Errorf(common.ErrUnknownMetric, metricTypeAsString)
	}

	return mType, nil
}

// UnmarshalJSON interface implementation.
func (gr *GetValueRequest) UnmarshalJSON(data []byte) (err error) {
	type ReqAlias GetValueRequest

	aliasValue := &struct {
		*ReqAlias
		MType string `json:"type"`
	}{
		ReqAlias: (*ReqAlias)(gr),
	}

	if err = json.Unmarshal(data, aliasValue); err != nil {
		return fmt.Errorf("unmarshal GetValueRequest: %w", err)
	}

	mType, parseError := parseType(aliasValue.MType)
	if parseError != nil {
		return fmt.Errorf("unmarshal GetValueRequest: %w", parseError)
	}

	gr.MType = mType
	return nil
}

// MetricDTO metric data contract.
type MetricDTO struct {
	Delta *int64            `json:"delta,omitempty"` // value for counter type
	Value *float64          `json:"value,omitempty"` // value for gauge type
	ID    string            `json:"id"`              // metric name
	MType common.MetricType `json:"type"`            // metric type
}

// MarshalJSON interface implementation.
func (m MetricDTO) MarshalJSON() ([]byte, error) {
	type DtoAlias MetricDTO

	aliasValue := struct {
		DtoAlias
		MType string `json:"type"`
	}{
		DtoAlias: DtoAlias(m),
		MType:    m.MType.String(),
	}

	bytes, err := json.Marshal(aliasValue)
	if err != nil {
		return nil, fmt.Errorf("marshal MetricDTO: %w", err)
	}
	return bytes, nil
}

// UnmarshalJSON interface implementation.
func (m *MetricDTO) UnmarshalJSON(data []byte) (err error) {
	type DtoAlias MetricDTO

	aliasValue := &struct {
		*DtoAlias
		MType string `json:"type"`
	}{
		DtoAlias: (*DtoAlias)(m),
	}

	if err = json.Unmarshal(data, aliasValue); err != nil {
		return fmt.Errorf("unmarshal MetricDTO: %w", err)
	}

	mType, parseError := parseType(aliasValue.MType)
	if parseError != nil {
		return fmt.Errorf("unmarshal MetricDTO type: %w", parseError)
	}

	m.MType = mType
	return nil
}
