package datatype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// NetworkState is one selectable network target for an icon group.
type NetworkState struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // wan | lan | custom
	Host string `json:"host"`
}

// NetworkConfig stores per-group network states and the active one.
type NetworkConfig struct {
	NetworkStates        []NetworkState `json:"networkStates"`
	ActiveNetworkStateId string         `json:"activeNetworkStateId"`
}

func DefaultNetworkConfig() NetworkConfig {
	return NetworkConfig{
		NetworkStates: []NetworkState{
			{Id: "wan", Name: "公网", Type: "wan", Host: ""},
			{Id: "lan", Name: "内网", Type: "lan", Host: ""},
		},
		ActiveNetworkStateId: "wan",
	}
}

func networkConfigBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, errors.New(fmt.Sprint("Failed to unmarshal NetworkConfig value:", value))
	}
}

func (j *NetworkConfig) Scan(value interface{}) error {
	bytes, err := networkConfigBytes(value)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		*j = DefaultNetworkConfig()
		return nil
	}
	cfg := NetworkConfig{}
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return err
	}
	if len(cfg.NetworkStates) == 0 {
		cfg = DefaultNetworkConfig()
	}
	if cfg.ActiveNetworkStateId == "" && len(cfg.NetworkStates) > 0 {
		cfg.ActiveNetworkStateId = cfg.NetworkStates[0].Id
	}
	*j = cfg
	return nil
}

func (j NetworkConfig) Value() (driver.Value, error) {
	if len(j.NetworkStates) == 0 {
		j = DefaultNetworkConfig()
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
