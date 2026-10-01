package release

import "testing"

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		latest  string
		current string
		want    bool
	}{
		{"1.0.2", "1.0.1", true},
		{"1.0.1", "1.0.0", true},
		{"1.0.1", "0.0.18", true},
		{"0.0.18", "1.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.2", false},
	}
	for _, tc := range cases {
		got := IsNewerVersion(tc.latest, tc.current)
		if got != tc.want {
			t.Fatalf("IsNewerVersion(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestSysmonApkAssetName(t *testing.T) {
	got := SysmonApkAssetName("1.0.43")
	want := "sysmon-1.0.43.apk"
	if got != want {
		t.Fatalf("SysmonApkAssetName(%q) = %q, want %q", "1.0.43", got, want)
	}
	gotFix := SysmonApkAssetName("v1.0.58-fix")
	wantFix := "sysmon-1.0.58-fix.apk"
	if gotFix != wantFix {
		t.Fatalf("SysmonApkAssetName(%q) = %q, want %q", "v1.0.58-fix", gotFix, wantFix)
	}
}

func TestAgentWindowsSetupAssetNamePreservesReleaseSuffix(t *testing.T) {
	got := AgentWindowsSetupAssetName("v1.0.58-fix")
	want := "system-monitor-agent_1.0.58-fix_setup.exe"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReleaseDownloadURLUsesTagAssetName(t *testing.T) {
	tag := "v1.0.58-fix"
	asset := AgentWindowsSetupAssetName(ReleaseAssetVersion(tag))
	url := ReleaseDownloadURL(tag, "1.0.58", asset)
	want := "https://github.com/pagbest154-cmd/system-monitor/releases/download/v1.0.58-fix/system-monitor-agent_1.0.58-fix_setup.exe"
	if url != want {
		t.Fatalf("ReleaseDownloadURL = %q, want %q", url, want)
	}
}

func TestPickLatestRelease(t *testing.T) {
	releases := []ghRelease{
		{TagName: "v0.0.18", HTMLURL: "https://example/0.0.18"},
		{TagName: "v1.0.1", HTMLURL: "https://example/1.0.1"},
		{TagName: "v1.0.0", HTMLURL: "https://example/1.0.0"},
		{TagName: "v1.0.2-rc1", Prerelease: true, HTMLURL: "https://example/rc"},
	}
	v, url, tag, ok := pickLatestRelease(releases)
	if !ok {
		t.Fatal("expected a release")
	}
	if v != "1.0.1" || tag != "v1.0.1" || url != "https://example/1.0.1" {
		t.Fatalf("got %q %q %q", v, tag, url)
	}
}
