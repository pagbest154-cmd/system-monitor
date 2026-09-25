package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/protocol"
)

type Transport struct {
	HubURL string
	Token  string
	client *http.Client
}

func NewTransport(hubURL, token string) *Transport {
	return &Transport{
		HubURL: strings.TrimRight(hubURL, "/"),
		Token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Transport) headers() map[string]string {
	h := map[string]string{"Content-Type": "application/json"}
	if t.Token != "" {
		h["Authorization"] = "Bearer " + t.Token
	}
	return h
}

func (t *Transport) PushReport(report *protocol.AgentReport) error {
	endpoint := fmt.Sprintf("%s/api/agents/%s/metrics", t.HubURL, report.AgentID)
	body, err := json.Marshal(report)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range t.headers() {
		req.Header.Set(k, v)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return WrapRequestError("push metrics", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return WrapHTTPError("push metrics", endpoint, resp.StatusCode, string(data))
	}
	return nil
}

func (t *Transport) SyncConfig(agentID string, configVersion int) (protocol.AgentConfigResponse, error) {
	endpoint := fmt.Sprintf("%s/api/agents/%s/config", t.HubURL, agentID)
	u, err := url.Parse(endpoint)
	if err != nil {
		return protocol.AgentConfigResponse{}, err
	}
	q := u.Query()
	q.Set("config_version", fmt.Sprintf("%d", configVersion))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return protocol.AgentConfigResponse{}, err
	}
	for k, v := range t.headers() {
		req.Header.Set(k, v)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return protocol.AgentConfigResponse{}, WrapRequestError("sync config", u.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return protocol.AgentConfigResponse{}, WrapHTTPError("sync config", u.String(), resp.StatusCode, string(data))
	}
	var out protocol.AgentConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return protocol.AgentConfigResponse{}, fmt.Errorf("sync config decode: %w", err)
	}
	return out, nil
}

func (t *Transport) FetchNotifyConfig(agentID string) (protocol.AgentNotifyResponse, error) {
	endpoint := fmt.Sprintf("%s/api/agents/%s/notify", t.HubURL, agentID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return protocol.AgentNotifyResponse{}, err
	}
	for k, v := range t.headers() {
		req.Header.Set(k, v)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return protocol.AgentNotifyResponse{}, WrapRequestError("fetch notify", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return protocol.AgentNotifyResponse{}, WrapHTTPError("fetch notify", endpoint, resp.StatusCode, string(data))
	}
	var out protocol.AgentNotifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return protocol.AgentNotifyResponse{}, fmt.Errorf("fetch notify decode: %w", err)
	}
	return out, nil
}

func (t *Transport) Close() {}
