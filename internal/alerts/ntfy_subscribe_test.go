package alerts

import "testing"

func TestNtfyWebSocketURL(t *testing.T) {
	cases := []struct {
		base  string
		topic string
		want  string
	}{
		{"https://ntfy.sh", "mytopic", "wss://ntfy.sh/mytopic/ws"},
		{"http://127.0.0.1:8090", "sysmon-homepc", "ws://127.0.0.1:8090/sysmon-homepc/ws"},
		{"wss://ntfy.example.com", "t", "wss://ntfy.example.com/t/ws"},
	}
	for _, tc := range cases {
		got := NtfyWebSocketURL(tc.base, tc.topic)
		if got != tc.want {
			t.Fatalf("NtfyWebSocketURL(%q, %q) = %q, want %q", tc.base, tc.topic, got, tc.want)
		}
	}
}

func TestParseNtfyWSMessage(t *testing.T) {
	title, body, ok := ParseNtfyWSMessage([]byte(`{"event":"open","topic":"t"}`))
	if ok {
		t.Fatal("open event should not parse as message")
	}
	title, body, ok = ParseNtfyWSMessage([]byte(`{"event":"message","title":"CPU","message":"85%"}`))
	if !ok || title != "CPU" || body != "85%" {
		t.Fatalf("unexpected parse: ok=%v title=%q body=%q", ok, title, body)
	}
}
