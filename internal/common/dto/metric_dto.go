package dto

import (
	"encoding/json"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
)

type GetValueRequest struct {
	ID    string            `json:"id"`   // имя метрики.
	MType common.MetricType `json:"type"` // параметр, принимающий значение gauge или counter.
}

func (gr *GetValueRequest) UnmarshalJSON(data []byte) (err error) {
	type ReqAlias GetValueRequest

	aliasValue := &struct {
		*ReqAlias
		MType string `json:"type"`
	}{
		ReqAlias: (*ReqAlias)(gr),
	}

	if err = json.Unmarshal(data, aliasValue); err != nil {
		return
	}

	mType, ok := common.StringToMetricType(aliasValue.MType)
	if !ok {
		return fmt.Errorf("unknown metric type: %s", aliasValue.MType)
	}

	if mType == common.UnknownMetricType {
		return fmt.Errorf("unknown metric type: %s", aliasValue.MType)
	}

	gr.MType = mType
	return nil
}

type MetricDTO struct {
	ID    string            `json:"id"`              // имя метрики.
	MType common.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter.
	Delta *int64            `json:"delta,omitempty"` // значение метрики в случае передачи counter.
	Value *float64          `json:"value,omitempty"` // значение метрики в случае передачи gauge.
}

func (m MetricDTO) MarshalJSON() ([]byte, error) {
	type DtoAlias MetricDTO

	aliasValue := struct {
		DtoAlias
		MType string `json:"type"`
	}{
		DtoAlias: DtoAlias(m),
		MType:    m.MType.String(),
	}

	return json.Marshal(aliasValue)
}

func (m *MetricDTO) UnmarshalJSON(data []byte) (err error) {
	type DtoAlias MetricDTO

	aliasValue := &struct {
		*DtoAlias
		MType string `json:"type"`
	}{
		DtoAlias: (*DtoAlias)(m),
	}

	if err = json.Unmarshal(data, aliasValue); err != nil {
		return
	}

	mType, ok := common.StringToMetricType(aliasValue.MType)
	if !ok {
		return fmt.Errorf("unknown metric type: %s", aliasValue.MType)
	}

	if mType == common.UnknownMetricType {
		return fmt.Errorf("unknown metric type: %s", aliasValue.MType)
	}

	m.MType = mType
	return nil
}
