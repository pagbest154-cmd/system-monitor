package protocol

import "fmt"

func PrefixedSensorID(agentID, sensorID string) string {
	return fmt.Sprintf("%s:%s", agentID, sensorID)
}

func StripAgentPrefix(agentID, fullID string) string {
	prefix := agentID + ":"
	if len(fullID) >= len(prefix) && fullID[:len(prefix)] == prefix {
		return fullID[len(prefix):]
	}
	return fullID
}

type MetricPoint struct {
	SensorID string   `json:"sensor_id"`
	TS       float64  `json:"ts"`
	Value    *float64 `json:"value"`
	Status   string   `json:"status"`
}

type SensorMeta struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Unit          string                 `json:"unit,omitempty"`
	Enabled       bool                   `json:"enabled"`
	IntervalSec   *int                   `json:"interval_sec,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty"`
	WarnAbove     *float64               `json:"warn_above,omitempty"`
	CriticalAbove *float64               `json:"critical_above,omitempty"`
}

type AgentReport struct {
	AgentID       string                 `json:"agent_id"`
	Hostname      string                 `json:"hostname,omitempty"`
	Metrics       []MetricPoint          `json:"metrics,omitempty"`
	System        map[string]interface{} `json:"system,omitempty"`
	ConfigVersion int                    `json:"config_version"`
	Sensors       []SensorMeta           `json:"sensors,omitempty"`
}

type SensorOverride struct {
	SensorID      string                 `json:"sensor_id"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	IntervalSec   *int                   `json:"interval_sec,omitempty"`
	WarnAbove     *float64               `json:"warn_above,omitempty"`
	CriticalAbove *float64               `json:"critical_above,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty"`
}

type AgentConfigResponse struct {
	Version   int              `json:"version"`
	Overrides []SensorOverride `json:"overrides,omitempty"`
}
