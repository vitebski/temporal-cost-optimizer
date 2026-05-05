package config

import "testing"

func TestLoadUsesSafeDefaults(t *testing.T) {
	cfg := Load(func(string) (string, bool) {
		return "", false
	})

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.Temporal.APIKey != "" {
		t.Fatalf("Temporal APIKey = %q, want empty", cfg.Temporal.APIKey)
	}
	if cfg.Temporal.Namespace != "" {
		t.Fatalf("Temporal namespace = %q, want empty", cfg.Temporal.Namespace)
	}
}

func TestLoadReadsTemporalSettingsFromEnv(t *testing.T) {
	env := map[string]string{
		"HTTP_ADDR":                 ":9090",
		"TEMPORAL_CLOUD_API_KEY":    "secret-token",
		"TEMPORAL_CLOUD_ACCOUNT_ID": "acct-123",
		"TEMPORAL_CLOUD_REGION":     "us-west1.gcp",
		"TEMPORAL_NAMESPACE":        "payments-prod",
	}

	cfg := Load(func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	})

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.Temporal.APIKey != "secret-token" {
		t.Fatalf("Temporal APIKey = %q, want secret-token", cfg.Temporal.APIKey)
	}
	if cfg.Temporal.AccountID != "acct-123" {
		t.Fatalf("Temporal account ID = %q, want acct-123", cfg.Temporal.AccountID)
	}
	if cfg.Temporal.Region != "us-west1.gcp" {
		t.Fatalf("Temporal region = %q, want us-west1.gcp", cfg.Temporal.Region)
	}
	if cfg.Temporal.Namespace != "payments-prod" {
		t.Fatalf("Temporal namespace = %q, want payments-prod", cfg.Temporal.Namespace)
	}
}
