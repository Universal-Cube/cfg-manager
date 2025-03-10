package config

import "sync"

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatYML  Format = "yml"
)

type Option func(*Manager)

type Manager struct {
	data          map[string]interface{}
	filePath      string
	fileFormat    Format
	caseSensitive bool
}

type ThreadSafeManager struct {
	mu      sync.RWMutex
	manager *Manager
}

type ConfigError struct {
	Operation string
	Key       string
	Err       error
}

func (e *ConfigError) Error() string {
	if e.Key != "" {
		return "config: " + e.Operation + " error with key '" + e.Key + "': " + e.Err.Error()
	}

	return "config: " + e.Operation + " error: " + e.Err.Error()
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}
