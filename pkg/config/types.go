package config

import (
	"fmt"
	"sync"
)

// Format represents the supported configuration file formats.
type Format string

// Supported configuration file formats.
const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatYML  Format = "yml"
)

// Option defines a function type for applying configuration options to a Manager.
type Option func(*Manager)

// Manager handles configuration data storage and retrieval operations.
// It supports loading from and saving to different file formats.
type Manager struct {
	data          map[string]interface{} // Configuration data
	filePath      string                 // Path to the configuration file
	fileFormat    Format                 // Format of the configuration file
	caseSensitive bool                   // Whether keys are case-sensitive

}

// ThreadSafeManager provides thread-safe access to a Manager instance.
type ThreadSafeManager struct {
	mu      *sync.RWMutex // Mutex for concurrent access control
	manager *Manager      // Underlying Manager instance
}

// ConfigError represents an error that occurred during configuration operations.
// It provides context about the operation and the key involved.
//
//goland:noinspection GoNameStartsWithPackageName
type ConfigError struct {
	Operation string // Operation being performed when the error occurred
	Key       string // Key being accessed (if applicable)
	Err       error  // Underlying error
}

// Error implements the error interface for ConfigError.
func (e *ConfigError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("config: %s error with key '%s': %v", e.Operation, e.Key, e.Err)
	}
	return fmt.Sprintf("config: %s error: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error for compatibility with errors.Is and errors.As.
func (e *ConfigError) Unwrap() error {
	return e.Err
}
