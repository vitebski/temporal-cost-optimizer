package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileUsesSafeDefaults(t *testing.T) {
	path := writeDotEnv(t, "")

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.Temporal.UsageAPIKey != "" {
		t.Fatalf("Temporal UsageAPIKey = %q, want empty", cfg.Temporal.UsageAPIKey)
	}
	if cfg.Temporal.NamespaceAPIKey != "" {
		t.Fatalf("Temporal NamespaceAPIKey = %q, want empty", cfg.Temporal.NamespaceAPIKey)
	}
	if cfg.Temporal.UsagePageSize != 100 {
		t.Fatalf("usage page size = %d, want 100", cfg.Temporal.UsagePageSize)
	}
}

func TestLoadFileReadsTemporalSettingsFromDotEnv(t *testing.T) {
	path := writeDotEnv(t, `
HTTP_ADDR=:9090
TEMPORAL_CLOUD_USAGE_API_KEY=usage-secret
TEMPORAL_CLOUD_NAMESPACE_API_KEY=namespace-secret
TEMPORAL_CLOUD_API_HOST_PORT=cloud.example.com:443
TEMPORAL_CLOUD_API_VERSION=v0.14.0
TEMPORAL_USAGE_PAGE_SIZE=250
`)

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.Temporal.UsageAPIKey != "usage-secret" {
		t.Fatalf("Temporal UsageAPIKey = %q, want usage-secret", cfg.Temporal.UsageAPIKey)
	}
	if cfg.Temporal.NamespaceAPIKey != "namespace-secret" {
		t.Fatalf("Temporal NamespaceAPIKey = %q, want namespace-secret", cfg.Temporal.NamespaceAPIKey)
	}
	if cfg.Temporal.APIHostPort != "cloud.example.com:443" {
		t.Fatalf("Temporal hostport = %q, want cloud.example.com:443", cfg.Temporal.APIHostPort)
	}
	if cfg.Temporal.APIVersion != "v0.14.0" {
		t.Fatalf("Temporal API version = %q, want v0.14.0", cfg.Temporal.APIVersion)
	}
	if cfg.Temporal.UsagePageSize != 250 {
		t.Fatalf("usage page size = %d, want 250", cfg.Temporal.UsagePageSize)
	}
}

func TestLoadFileRejectsMalformedLine(t *testing.T) {
	path := writeDotEnv(t, "NOT_VALID\n")

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("LoadFile returned nil error, want malformed line error")
	}
}

func writeDotEnv(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	return path
}
