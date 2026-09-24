package sysinfo

import (
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

var ignoreFSTypes = map[string]bool{
	"tmpfs": true, "devtmpfs": true, "squashfs": true, "overlay": true,
	"proc": true, "sysfs": true, "devfs": true, "autofs": true,
	"cgroup": true, "cgroup2": true, "pstore": true, "bpf": true,
	"tracefs": true, "debugfs": true, "securityfs": true, "configfs": true,
	"fusectl": true, "mqueue": true, "hugetlbfs": true,
}

func bytesToGB(value uint64) float64 {
	return round1(float64(value) / (1024 * 1024 * 1024))
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func GetSystemInfo() map[string]interface{} {
	vm, _ := mem.VirtualMemory()
	swap, _ := mem.SwapMemory()
	partitions := getPartitions()
	bootTime, _ := host.BootTime()
	hostname, _ := os.Hostname()
	info, _ := host.Info()

	cpuSection := getCPUSection()
	gpus := getGPUs()

	return map[string]interface{}{
		"hostname":       hostname,
		"os":             getOSLabel(info),
		"os_version":     runtime.Version(),
		"platform":       runtime.GOOS,
		"architecture":   runtime.GOARCH,
		"python_version": runtime.Version(),
		"cpu":            cpuSection,
		"memory": map[string]interface{}{
			"total_gb":     bytesToGB(vm.Total),
			"used_gb":      bytesToGB(vm.Used),
			"available_gb": bytesToGB(vm.Available),
			"free_gb":      bytesToGB(vm.Free),
			"percent":      round1(vm.UsedPercent),
		},
		"swap": map[string]interface{}{
			"total_gb": bytesToGB(swap.Total),
			"used_gb":  bytesToGB(swap.Used),
			"free_gb":  bytesToGB(swap.Free),
			"percent":  round1(swap.UsedPercent),
		},
		"physical_drives": getPhysicalDrives(),
		"partitions":      partitions,
		"disks":           partitions,
		"gpus":            gpus,
		"network":         getNetworkInterfaces(),
		"battery":         getBattery(),
		"boot_time":       float64(bootTime),
		"uptime_sec":      int(time.Now().Unix()) - int(bootTime),
	}
}

func getOSLabel(info *host.InfoStat) string {
	if info == nil {
		return runtime.GOOS
	}
	return info.Platform + " " + info.PlatformVersion
}

func getCPUSection() map[string]interface{} {
	infos, _ := cpu.Info()
	name := "Unknown CPU"
	if len(infos) > 0 {
		name = infos[0].ModelName
	}
	physical, _ := cpu.Counts(false)
	logical, _ := cpu.Counts(true)
	perCore, _ := cpu.Percent(0, true)
	var percent float64
	if len(perCore) > 0 {
		sum := 0.0
		for _, v := range perCore {
			sum += v
		}
		percent = round1(sum / float64(len(perCore)))
	}
	freqs, _ := cpu.Info()
	var freqMHz *float64
	if len(freqs) > 0 && freqs[0].Mhz > 0 {
		f := round1(freqs[0].Mhz)
		freqMHz = &f
	}
	perCoreRounded := make([]float64, len(perCore))
	for i, v := range perCore {
		perCoreRounded[i] = round1(v)
	}
	return map[string]interface{}{
		"name":              name,
		"cores_physical":    physical,
		"cores_logical":     logical,
		"freq_mhz":          freqMHz,
		"freq_min_mhz":      nil,
		"freq_max_mhz":      nil,
		"percent":           percent,
		"per_core_percent":    perCoreRounded,
		"per_core_freq_mhz":   []interface{}{},
	}
}

func getPartitions() []map[string]interface{} {
	all := runtime.GOOS != "windows"
	parts, err := disk.Partitions(all)
	if err != nil {
		parts, _ = disk.Partitions(false)
	}
	seen := map[string]bool{}
	result := make([]map[string]interface{}, 0)
	for _, part := range parts {
		if seen[part.Mountpoint] || ignoreFSTypes[part.Fstype] {
			continue
		}
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		seen[part.Mountpoint] = true
		result = append(result, map[string]interface{}{
			"device":     part.Device,
			"mountpoint": part.Mountpoint,
			"fstype":     part.Fstype,
			"opts":       part.Opts,
			"media_type": nil,
			"total_gb":   bytesToGB(usage.Total),
			"used_gb":    bytesToGB(usage.Used),
			"free_gb":    bytesToGB(usage.Free),
			"percent":    round1(usage.UsedPercent),
		})
	}
	return result
}

func getNetworkInterfaces() []map[string]interface{} {
	ifaces, _ := net.Interfaces()
	io, _ := net.IOCounters(true)
	ioMap := map[string]net.IOCountersStat{}
	for _, item := range io {
		ioMap[item.Name] = item
	}
	result := make([]map[string]interface{}, 0)
	for _, iface := range ifaces {
		if iface.Name == "lo" || iface.Name == "Loopback Pseudo-Interface 1" {
			continue
		}
		ipv4 := make([]string, 0)
		ipv6 := make([]string, 0)
		mac := iface.HardwareAddr
		for _, addr := range iface.Addrs {
			s := addr.Addr
			if len(s) > 0 && s[0] != ':' {
				if len(s) > 15 {
					ipv6 = append(ipv6, s)
				} else {
					ipv4 = append(ipv4, s)
				}
			}
		}
		counters := ioMap[iface.Name]
		result = append(result, map[string]interface{}{
			"name":            iface.Name,
			"ipv4":            ipv4,
			"ipv6":            ipv6,
			"mac":             mac,
			"is_up":           true,
			"speed_mbps":      nil,
			"bytes_sent_gb":     bytesToGB(counters.BytesSent),
			"bytes_recv_gb":     bytesToGB(counters.BytesRecv),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	return result
}

func getBattery() interface{} {
	return nil
}

func getPhysicalDrives() []map[string]interface{} {
	return []map[string]interface{}{}
}

func getGPUs() []map[string]interface{} {
	return getNvidiaGPUs()
}
