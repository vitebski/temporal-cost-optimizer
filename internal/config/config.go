package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr string
	Temporal TemporalConfig
}

type TemporalConfig struct {
	UsageAPIKey     string
	NamespaceAPIKey string
	APIHostPort     string
	APIVersion      string
	UsagePageSize   int32
}

func LoadFile(path string) (Config, error) {
	values, err := readDotEnv(path)
	if err != nil {
		return Config{}, err
	}

	usagePageSize, err := getInt32(values, "TEMPORAL_USAGE_PAGE_SIZE", 100)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr: getValue(values, "HTTP_ADDR", ":8080"),
		Temporal: TemporalConfig{
			UsageAPIKey:     getValue(values, "TEMPORAL_CLOUD_USAGE_API_KEY", ""),
			NamespaceAPIKey: getValue(values, "TEMPORAL_CLOUD_NAMESPACE_API_KEY", ""),
			APIHostPort:     getValue(values, "TEMPORAL_CLOUD_API_HOST_PORT", ""),
			APIVersion:      getValue(values, "TEMPORAL_CLOUD_API_VERSION", ""),
			UsagePageSize:   usagePageSize,
		},
	}, nil
}

func readDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .env file %q: %w", path, err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("parse .env file %q line %d: expected KEY=value", path, lineNumber)
		}

		values[key] = trimValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .env file %q: %w", path, err)
	}

	return values, nil
}

func trimValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}
		if value[0] == '\'' && value[len(value)-1] == '\'' {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func getValue(values map[string]string, key string, fallback string) string {
	value, ok := values[key]
	if !ok || value == "" {
		return fallback
	}
	return value
}

func getInt32(values map[string]string, key string, fallback int32) (int32, error) {
	value, ok := values[key]
	if !ok || value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return int32(parsed), nil
}
