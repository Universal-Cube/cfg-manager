package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func WithCaseSensitive(sensitive bool) Option {
	return func(m *Manager) {
		m.caseSensitive = sensitive
	}
}

func New(options ...Option) *Manager {
	m := &Manager{
		data:          make(map[string]interface{}),
		caseSensitive: true,
	}

	for _, option := range options {
		option(m)
	}

	return m
}

func (m *Manager) LoadFile(filePath string) error {
	resolvedPath, err := resolvePath(filePath)
	if err != nil {
		return &ConfigError{
			Operation: "resolve path",
			Err:       err,
		}
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return &ConfigError{
			Operation: "stat file",
			Err:       err,
		}
	}

	if !info.Mode().IsRegular() {
		return &ConfigError{
			Operation: "check file",
			Err:       fmt.Errorf("file '%s' is not a regular file", filePath),
		}
	}

	format, err := detectFileFormat(resolvedPath)
	if err != nil {
		return &ConfigError{
			Operation: "detect file format",
			Err:       err,
		}
	}

	file, err := os.Open(resolvedPath)
	if err != nil {
		return &ConfigError{
			Operation: "open file",
			Err:       err,
		}
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return m.Load(file, format)
}

func (m *Manager) Load(r io.Reader, format Format) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return &ConfigError{
			Operation: "read file content",
			Err:       err,
		}
	}

	switch format {
	case FormatJSON:
		err = json.Unmarshal(content, &m.data)
	case FormatYAML, FormatYML:
		var data interface{}
		err = yaml.Unmarshal(content, &data)
		if err == nil {
			if mapData, ok := data.(map[string]interface{}); ok {
				m.data = mapData
			} else if mapData, ok := data.(map[interface{}]interface{}); ok {
				transformed := transformMapKeys(mapData)
				if strMap, ok := transformed.(map[string]interface{}); ok {
					m.data = strMap
				} else {
					err = errors.New("failed to convert YAML data to map[string]interface{}")
				}
			} else {
				err = errors.New("unexpected YAML structure")
			}
		}
	default:
		err = fmt.Errorf("unsupported file format: %s", format)
	}

	if err != nil {
		return &ConfigError{
			Operation: "parse",
			Err:       err,
		}
	}

	m.fileFormat = format
	return nil
}
