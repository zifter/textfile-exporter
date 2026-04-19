package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func resetViper() {
	viper.Reset()
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetViper()

	conf, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.ServeAddr != ":8080" {
		t.Errorf("ServeAddr: got %q, want %q", conf.ServeAddr, ":8080")
	}
	if conf.MetricsFilePath != "metrics.txt" {
		t.Errorf("MetricsFilePath: got %q, want %q", conf.MetricsFilePath, "metrics.txt")
	}
	if conf.MetricsEndpoint != "/metrics" {
		t.Errorf("MetricsEndpoint: got %q, want %q", conf.MetricsEndpoint, "/metrics")
	}
	if conf.RefreshInterval != 0 {
		t.Errorf("RefreshInterval: got %v, want 0", conf.RefreshInterval)
	}
	if conf.LogOutput != "stdout" {
		t.Errorf("LogOutput: got %q, want %q", conf.LogOutput, "stdout")
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	resetViper()
	t.Setenv("SERVE_ADDR", ":9090")
	t.Setenv("METRICS_FILE_PATH", "/tmp/custom.txt")
	t.Setenv("REFRESH_INTERVAL", "30s")

	conf, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conf.ServeAddr != ":9090" {
		t.Errorf("ServeAddr: got %q, want %q", conf.ServeAddr, ":9090")
	}
	if conf.MetricsFilePath != "/tmp/custom.txt" {
		t.Errorf("MetricsFilePath: got %q, want %q", conf.MetricsFilePath, "/tmp/custom.txt")
	}
	if conf.RefreshInterval != 30*time.Second {
		t.Errorf("RefreshInterval: got %v, want 30s", conf.RefreshInterval)
	}
}

func TestMetricsExporter_LoadFromFile(t *testing.T) {
	content := "# HELP foo A metric\n# TYPE foo gauge\nfoo 1\n"
	f, err := os.CreateTemp(t.TempDir(), "metrics*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()

	exp := &MetricsExporter{}
	if err := exp.loadFromFile(f.Name()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(exp.content) != content {
		t.Errorf("content mismatch: got %q, want %q", exp.content, content)
	}
}

func TestMetricsExporter_LoadFromFile_NotFound(t *testing.T) {
	exp := &MetricsExporter{}
	err := exp.loadFromFile("/nonexistent/metrics.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMetricsExporter_Handler(t *testing.T) {
	content := "foo 1\n"
	exp := &MetricsExporter{content: []byte(content)}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	exp.handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; version=0.0.4" {
		t.Errorf("Content-Type: got %q", ct)
	}
	if body := w.Body.String(); body != content {
		t.Errorf("body: got %q, want %q", body, content)
	}
}

func TestOkHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	okHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if body := w.Body.String(); body != "OK" {
		t.Errorf("body: got %q, want %q", body, "OK")
	}
}
