package mdadmdiscovery

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

const mdstatPath = "/proc/mdstat"

// Array describes one Linux md RAID array from /proc/mdstat.
type Array struct {
	Device        string
	RaidLevel     string
	State         string
	ActiveDevices int
	TotalDevices  int
	FailedDevices int
	CheckProgress *float64
}

var (
	mdHeaderRe = regexp.MustCompile(`^(\S+)\s*:\s*(\S+)(?:\s+(\S+))?`)
	countRe    = regexp.MustCompile(`\[(\d+)/(\d+)\]`)
	stateRe    = regexp.MustCompile(`\[([U_]+)\]`)
	checkRe    = regexp.MustCompile(`(?:check|recovery|resync|reshape)\s*=\s*([\d.]+)%`)
)

// ParseMdstat parses the contents of /proc/mdstat.
func ParseMdstat(content string) []Array {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	var out []Array
	var current *Array
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Personalities") || strings.HasPrefix(trimmed, "unused devices") {
			continue
		}
		if m := mdHeaderRe.FindStringSubmatch(trimmed); m != nil {
			if current != nil {
				out = append(out, finalizeArray(*current))
			}
			raidLevel := ""
			if m[3] != "" && strings.HasPrefix(m[3], "raid") {
				raidLevel = m[3]
			}
			current = &Array{
				Device:    m[1],
				State:     m[2],
				RaidLevel: raidLevel,
			}
			continue
		}
		if current == nil {
			continue
		}
		if current.RaidLevel == "" {
			if idx := strings.Index(trimmed, "raid"); idx >= 0 {
				fields := strings.Fields(trimmed[idx:])
				if len(fields) > 0 && strings.HasPrefix(fields[0], "raid") {
					current.RaidLevel = fields[0]
				}
			}
		}
		if cm := countRe.FindStringSubmatch(trimmed); cm != nil {
			active, _ := strconv.Atoi(cm[1])
			total, _ := strconv.Atoi(cm[2])
			current.ActiveDevices = active
			current.TotalDevices = total
			if total > 0 {
				current.FailedDevices = total - active
			}
		}
		if sm := stateRe.FindStringSubmatch(trimmed); sm != nil {
			failed := 0
			for _, ch := range sm[1] {
				if ch == '_' {
					failed++
				}
			}
			if failed > current.FailedDevices {
				current.FailedDevices = failed
			}
		}
		if pm := checkRe.FindStringSubmatch(trimmed); pm != nil {
			pct, err := strconv.ParseFloat(pm[1], 64)
			if err == nil {
				current.CheckProgress = &pct
			}
		}
	}
	if current != nil {
		out = append(out, finalizeArray(*current))
	}
	return out
}

func finalizeArray(a Array) Array {
	if a.TotalDevices > 0 && a.ActiveDevices < a.TotalDevices {
		if a.State == "active" || a.State == "clean" {
			a.State = "degraded"
		}
	}
	if a.FailedDevices > 0 && a.State != "degraded" {
		a.State = "degraded"
	}
	return a
}

// ReadMdstat reads and parses /proc/mdstat. Missing file yields nil slice, no error.
func ReadMdstat() []Array {
	data, err := os.ReadFile(mdstatPath)
	if err != nil {
		return nil
	}
	return ParseMdstat(string(data))
}

// DeviceToSensorID returns mdadm_md0 for device md0.
func DeviceToSensorID(device string) string {
	return "mdadm_" + strings.TrimSpace(device)
}
