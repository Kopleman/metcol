package postgres

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	// In seconds. Default 5 sec
	PingInterval         int   `json:"POSTGRES_PING_INTERVAL" default:"5"`
	MaxConns             int32 `json:"POSTGRES_MAX_CONNS"`
	MinConns             int32 `json:"POSTGRES_MIN_CONNS"`
	PreferSimpleProtocol bool  `json:"POSTGRES_PREFER_SIMPLE_PROTOCOL"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.PingInterval, validation.Required, validation.Min(1)),
	)
}

// GetMaxConns return max conns
func (c Config) GetMaxConns() int32 {
	return c.MaxConns
}

// GetMinConns return min conns
func (c Config) GetMinConns() int32 {
	return c.MinConns
}

// GetPreferSimpleProtocol return PreferSimpleProtocol option
func (c Config) GetPreferSimpleProtocol() bool {
	return c.PreferSimpleProtocol
}
