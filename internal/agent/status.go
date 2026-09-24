package agent

import (
	"encoding/json"
	"os"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

type Status struct {
	Connected        bool    `json:"connected"`
	AgentID          string  `json:"agent_id"`
	HubURL           string  `json:"hub_url"`
	Hostname         string  `json:"hostname"`
	LastSuccessTS    float64 `json:"last_success_ts,omitempty"`
	LastError        *string `json:"last_error,omitempty"`
	LastErrorTS      float64 `json:"last_error_ts,omitempty"`
	MetricsCount     int     `json:"metrics_count,omitempty"`
	UpdateAvailable  bool    `json:"update_available,omitempty"`
	LatestVersion    string  `json:"latest_version,omitempty"`
	ReleaseURL       string  `json:"release_url,omitempty"`
	UpdateCheckedAt  float64 `json:"update_checked_at,omitempty"`
}

func (s Status) IsStale() bool {
	if s.LastSuccessTS == 0 {
		return true
	}
	return time.Now().Unix()-int64(s.LastSuccessTS) > 90
}

func WriteStatus(status Status) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(paths.AgentStatusFile, data, 0o644)
}

func ReadStatus() (*Status, error) {
	data, err := os.ReadFile(paths.AgentStatusFile)
	if err != nil {
		return nil, err
	}
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
