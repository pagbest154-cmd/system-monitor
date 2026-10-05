package mdadmdiscovery

import (
	"testing"
)

func TestParseMdstat_empty(t *testing.T) {
	if ParseMdstat("") != nil {
		t.Fatal("expected nil for empty")
	}
	if ParseMdstat("Personalities : [raid1]\n") != nil {
		t.Fatal("expected nil for personalities only")
	}
}

func TestParseMdstat_activeRaid1(t *testing.T) {
	content := `Personalities : [raid1] [raid0]
md0 : active raid1 sda1[0] sdb1[1]
      2097088 blocks super 1.2 [2/2] [UU]
`
	arrays := ParseMdstat(content)
	if len(arrays) != 1 {
		t.Fatalf("expected 1 array, got %d", len(arrays))
	}
	a := arrays[0]
	if a.Device != "md0" || a.RaidLevel != "raid1" || a.State != "active" {
		t.Fatalf("unexpected array: %+v", a)
	}
	if a.ActiveDevices != 2 || a.TotalDevices != 2 || a.FailedDevices != 0 {
		t.Fatalf("devices: %+v", a)
	}
	if a.CheckProgress != nil {
		t.Fatal("expected no check progress")
	}
}

func TestParseMdstat_degraded(t *testing.T) {
	content := `md127 : active raid5 sdc1[0] sdd1[1] sde1[2]
      5857286144 blocks super 1.2 level 5, 512k chunk, algorithm 2 [2/3] [UU_]
`
	arrays := ParseMdstat(content)
	if len(arrays) != 1 {
		t.Fatalf("expected 1 array, got %d", len(arrays))
	}
	a := arrays[0]
	if a.State != "degraded" {
		t.Fatalf("expected degraded, got %s", a.State)
	}
	if a.FailedDevices < 1 {
		t.Fatalf("expected failed devices, got %d", a.FailedDevices)
	}
}

func TestParseMdstat_checkProgress(t *testing.T) {
	content := `md0 : active raid1 sda1[0] sdb1[1]
      2097088 blocks super 1.2 [2/2] [UU]
      [=>...................]  check = 12.5% (1048544/8388608)
`
	arrays := ParseMdstat(content)
	if len(arrays) != 1 {
		t.Fatalf("expected 1 array, got %d", len(arrays))
	}
	if arrays[0].CheckProgress == nil || *arrays[0].CheckProgress != 12.5 {
		t.Fatalf("check progress: %+v", arrays[0].CheckProgress)
	}
}

func TestDeviceToSensorID(t *testing.T) {
	if DeviceToSensorID("md0") != "mdadm_md0" {
		t.Fatal("unexpected sensor id")
	}
}
