package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

const (
	GitHubRepo              = "pagbest154-cmd/system-monitor"
	ReleasesLatestPage      = "https://github.com/" + GitHubRepo + "/releases/latest"
	ReleasesAPI             = "https://api.github.com/repos/" + GitHubRepo + "/releases?per_page=100"
	DefaultCheckInterval    = 24 * 60 * 60
	HubVersionCheckInterval = 60 * 60 // hub footer: refresh at most once per hour
	UpToDateRecheckInterval = 60 * 60 // re-check GitHub if cache claims "up to date"
)

type ReleaseCheckResult struct {
	CurrentVersion  string  `json:"current_version"`
	LatestVersion   string  `json:"latest_version"`
	UpdateAvailable bool    `json:"update_available"`
	ReleaseURL      string  `json:"release_url"`
	DownloadURL     *string `json:"download_url"`
	CheckedAt       float64 `json:"checked_at"`
	Error           *string `json:"error"`
}

type ghRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	HTMLURL    string `json:"html_url"`
}

func NormalizeVersion(v string) string {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}
	return v
}

func VersionKey(v string) []int {
	parts := strings.Split(NormalizeVersion(v), ".")
	out := make([]int, 0, len(parts))
	for _, piece := range parts {
		n := 0
		fmt.Sscanf(piece, "%d", &n)
		out = append(out, n)
	}
	return out
}

func IsNewerVersion(latest, current string) bool {
	l := VersionKey(latest)
	c := VersionKey(current)
	maxLen := len(l)
	if len(c) > maxLen {
		maxLen = len(c)
	}
	for i := 0; i < maxLen; i++ {
		lv, cv := 0, 0
		if i < len(l) {
			lv = l[i]
		}
		if i < len(c) {
			cv = c[i]
		}
		if lv > cv {
			return true
		}
		if lv < cv {
			return false
		}
	}
	return false
}

func ReleaseDownloadURL(tag, latestVersion, assetName string) string {
	tagName := tag
	if !strings.HasPrefix(tagName, "v") {
		tagName = "v" + latestVersion
	}
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", GitHubRepo, tagName, assetName)
}

func readCache(path string) (*ReleaseCheckResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result ReleaseCheckResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func writeCache(path string, result ReleaseCheckResult) error {
	if err := os.MkdirAll(strings.TrimSuffix(path, "/"+strings.Split(path, "/")[len(strings.Split(path, "/"))-1]), 0o755); err != nil {
		// fallback mkdir parent
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func parseReleaseTag(finalURL, body string) string {
	re := regexp.MustCompile(`/releases/tag/(v?[^/?#]+)`)
	if m := re.FindStringSubmatch(finalURL); len(m) > 1 {
		return m[1]
	}
	re2 := regexp.MustCompile(`href="[^"]*/releases/tag/(v?[^"/?#]+)"`)
	if m := re2.FindStringSubmatch(body); len(m) > 1 {
		return m[1]
	}
	return ""
}

func fetchReleasesFromAPI(userAgent string) ([]ghRelease, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, ReleasesAPI, nil)
	if err != nil {
		return nil, err
	}
	if userAgent == "" {
		userAgent = "system-monitor"
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func pickLatestRelease(releases []ghRelease) (latestVersion, releaseURL, tag string, ok bool) {
	for _, rel := range releases {
		if rel.Draft || rel.Prerelease {
			continue
		}
		v := NormalizeVersion(rel.TagName)
		if v == "" {
			continue
		}
		if !ok || IsNewerVersion(v, latestVersion) {
			latestVersion = v
			tag = rel.TagName
			releaseURL = rel.HTMLURL
			ok = true
		}
	}
	return latestVersion, releaseURL, tag, ok
}

func fetchLatestReleaseHTML(userAgent string) (latestVersion, releaseURL, tag string, err error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, ReleasesLatestPage, nil)
	if err != nil {
		return "", "", "", err
	}
	if userAgent == "" {
		userAgent = "system-monitor"
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	finalURL := resp.Request.URL.String()
	tag = parseReleaseTag(finalURL, body)
	if tag == "" {
		return "", "", "", fmt.Errorf("cannot parse release version")
	}
	latestVersion = NormalizeVersion(tag)
	if latestVersion == "" {
		return "", "", "", fmt.Errorf("empty release version")
	}
	return latestVersion, finalURL, tag, nil
}

func FetchLatestRelease(userAgent string) (latestVersion, releaseURL, tag string, err error) {
	releases, err := fetchReleasesFromAPI(userAgent)
	if err == nil {
		v, url, t, ok := pickLatestRelease(releases)
		if ok {
			return v, url, t, nil
		}
	}

	// Fallback for API rate limits or empty list.
	return fetchLatestReleaseHTML(userAgent)
}

type AssetNameFunc func(version string) string

func AgentDebAssetName(v string) string {
	return fmt.Sprintf("system-monitor-agent_%s-1_amd64.deb", NormalizeVersion(v))
}

func AgentWindowsSetupAssetName(v string) string {
	return fmt.Sprintf("system-monitor-agent_%s_setup.exe", NormalizeVersion(v))
}

func CheckReleaseUpdates(currentVersion, cachePath string, force bool, userAgent string, checkInterval int, assetName AssetNameFunc) ReleaseCheckResult {
	current := NormalizeVersion(currentVersion)
	now := float64(time.Now().UnixNano()) / 1e9
	if checkInterval <= 0 {
		checkInterval = DefaultCheckInterval
	}
	var cached *ReleaseCheckResult
	if cachePath != "" {
		cached, _ = readCache(cachePath)
	}
	if !force && cached != nil {
		age := now - cached.CheckedAt
		cachedAtVersion := NormalizeVersion(cached.CurrentVersion)
		cached.CurrentVersion = current
		cached.UpdateAvailable = IsNewerVersion(cached.LatestVersion, current)
		cached.Error = nil
		needsRefresh := age >= float64(checkInterval)
		// Hub/package was upgraded since the last check — cached latest is stale.
		if cachedAtVersion != current {
			needsRefresh = true
		}
		// Cache may say "up to date" while new releases were published after the last check.
		if !needsRefresh && !cached.UpdateAvailable && age >= float64(UpToDateRecheckInterval) {
			needsRefresh = true
		}
		if !needsRefresh {
			return *cached
		}
	}
	latestVersion, releaseURL, tag, err := FetchLatestRelease(userAgent)
	if err != nil {
		if cached != nil && !force {
			cached.CurrentVersion = current
			cached.UpdateAvailable = IsNewerVersion(cached.LatestVersion, current)
			cached.Error = nil
			return *cached
		}
		errStr := err.Error()
		return ReleaseCheckResult{
			CurrentVersion: current, LatestVersion: current, UpdateAvailable: false,
			ReleaseURL: ReleasesLatestPage, CheckedAt: now, Error: &errStr,
		}
	}
	var downloadURL *string
	if assetName != nil {
		url := ReleaseDownloadURL(tag, latestVersion, assetName(latestVersion))
		downloadURL = &url
	}
	result := ReleaseCheckResult{
		CurrentVersion:  current,
		LatestVersion:   latestVersion,
		UpdateAvailable: IsNewerVersion(latestVersion, current),
		ReleaseURL:      releaseURL,
		DownloadURL:     downloadURL,
		CheckedAt:       now,
	}
	if cachePath != "" {
		_ = writeCache(cachePath, result)
	}
	return result
}

func InstalledVersion() string {
	return version.Version
}
