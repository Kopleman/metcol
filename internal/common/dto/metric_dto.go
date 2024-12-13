package dto

import (
	"encoding/json"
	"fmt"

	"github.com/Kopleman/metcol/internal/common"
)

type MetricDto struct {
	ID    string            `json:"id"`              // имя метрики.
	MType common.MetricType `json:"type"`            // параметр, принимающий значение gauge или counter.
	Delta *string           `json:"delta,omitempty"` // значение метрики в случае передачи counter.
	Value *string           `json:"value,omitempty"` // значение метрики в случае передачи gauge.
}

func (m MetricDto) MarshalJSON() ([]byte, error) {
	// чтобы избежать рекурсии при json.Marshal, объявляем новый тип
	type DtoAlias MetricDto

	aliasValue := struct {
		DtoAlias
		MType string `json:"type"`
	}{
		DtoAlias: DtoAlias(m),
		MType:    m.MType.String(),
	}

	return json.Marshal(aliasValue) // вызываем стандартный Marshal
}

func (m *MetricDto) UnmarshalJSON(data []byte) (err error) {
	type DtoAlias MetricDto

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
