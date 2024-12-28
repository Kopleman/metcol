// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package pgxstore

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

func (e *MetricType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MetricType(s)
	case string:
		*e = MetricType(s)
	default:
		return fmt.Errorf("unsupported scan type for MetricType: %T", src)
	}
	return nil
}

type NullMetricType struct {
	MetricType MetricType `json:"metric_type"`
	Valid      bool       `json:"valid"` // Valid is true if MetricType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMetricType) Scan(value interface{}) error {
	if value == nil {
		ns.MetricType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MetricType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMetricType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MetricType), nil
}

func (e MetricType) Valid() bool {
	switch e {
	case MetricTypeGauge,
		MetricTypeCounter:
		return true
	}
	return false
}

func AllMetricTypeValues() []MetricType {
	return []MetricType{
		MetricTypeGauge,
		MetricTypeCounter,
	}
}

type Metric struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Type      MetricType `db:"type" json:"type"`
	Value     *float64   `db:"value" json:"value"`
	Delta     *int64     `db:"delta" json:"delta"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at"`
}
