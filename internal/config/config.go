package config

import "os"

type Config struct {
	HTTPAddr string
	Temporal TemporalConfig
}

type TemporalConfig struct {
	APIKey    string
	AccountID string
	Region    string
	Namespace string
}

func Load(lookup func(string) (string, bool)) Config {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	return Config{
		HTTPAddr: getEnv(lookup, "HTTP_ADDR", ":8080"),
		Temporal: TemporalConfig{
			APIKey:    getEnv(lookup, "TEMPORAL_CLOUD_API_KEY", ""),
			AccountID: getEnv(lookup, "TEMPORAL_CLOUD_ACCOUNT_ID", ""),
			Region:    getEnv(lookup, "TEMPORAL_CLOUD_REGION", ""),
			Namespace: getEnv(lookup, "TEMPORAL_NAMESPACE", ""),
		},
	}
}

func getEnv(lookup func(string) (string, bool), key string, fallback string) string {
	value, ok := lookup(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}
