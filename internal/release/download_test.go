package release

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("installer-payload"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "setup.exe")
	if err := DownloadFile(srv.URL, dest, "test"); err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != "installer-payload" {
		t.Fatalf("unexpected payload: %q", data)
	}
}

func TestDownloadFileHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "setup.exe")
	if err := DownloadFile(srv.URL, dest, "test"); err == nil {
		t.Fatal("expected error for 404")
	}
}
