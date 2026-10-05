package mdadmdiscovery

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

const mdstatPath = "/proc/mdstat"

// MemberDevice is one block device in an md array (from mdstat header + [UU_] line).
type MemberDevice struct {
	Name  string
	Slot  int
	State string // active | failed
}

// Array describes one Linux md RAID array from /proc/mdstat.
type Array struct {
	Device        string
	RaidLevel     string
	State         string
	ActiveDevices int
	TotalDevices  int
	FailedDevices int
	CheckProgress *float64
	Devices       []MemberDevice
}

var (
	mdHeaderRe = regexp.MustCompile(`^(\S+)\s*:\s*(\S+)(?:\s+(\S+))?`)
	deviceRe   = regexp.MustCompile(`(\S+)\[(\d+)\]`)
	countRe    = regexp.MustCompile(`\[(\d+)/(\d+)\]`)
	stateRe    = regexp.MustCompile(`\[([U_]+)\]`)
	checkRe    = regexp.MustCompile(`(?:check|recovery|resync|reshape)\s*=\s*([\d.]+)%`)
)

// HealthMetricValue maps array health to a scalar for alerts: 0 ok, 1 degraded, 2+ failed disk(s).
func HealthMetricValue(a Array) float64 {
	if a.FailedDevices > 0 {
		return 2
	}
	if a.State == "degraded" {
		return 1
	}
	return 0
}

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
				Devices:   parseMemberDevices(trimmed),
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
			for i, ch := range sm[1] {
				state := "active"
				if ch == '_' {
					state = "failed"
					failed++
				}
				if i < len(current.Devices) {
					current.Devices[i].State = state
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

func parseMemberDevices(line string) []MemberDevice {
	members := make([]MemberDevice, 0)
	for _, m := range deviceRe.FindAllStringSubmatch(line, -1) {
		slot, _ := strconv.Atoi(m[2])
		members = append(members, MemberDevice{Name: m[1], Slot: slot, State: "active"})
	}
	return members
}
