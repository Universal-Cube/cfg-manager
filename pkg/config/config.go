package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
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

func (m *Manager) Get(key string) (interface{}, error) {
	if key == "" {
		return m.data, nil
	}

	parentMap, lastKey, err := getNestedMap(m.data, key, m.caseSensitive)
	if err != nil {
		return nil, &ConfigError{
			Operation: "get nested map",
			Key:       key,
			Err:       err,
		}
	}

	if !m.caseSensitive {
		for key := range parentMap {
			if strings.EqualFold(key, lastKey) {
				lastKey = key
				break
			}
		}
	}

	value, exists := parentMap[lastKey]
	if !exists {
		return nil, &ConfigError{
			Operation: "get value",
			Key:       key,
			Err:       fmt.Errorf("key '%s' not found", key),
		}
	}

	return value, nil
}

func (m *Manager) GetString(key string) (string, error) {
	value, err := m.Get(key)
	if err != nil {
		return "", err
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (m *Manager) GetBool(key string) (bool, error) {
	value, err := m.Get(key)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "1" || lower == "yes" || lower == "y" || lower == "on" {
			return true, nil
		} else if lower == "false" || lower == "0" || lower == "no" || lower == "n" || lower == "off" {
			return false, nil
		}
		return false, &ConfigError{
			Operation: "convert",
			Key:       key,
			Err:       errors.New("cannot convert string to bool"),
		}
	default:
		return false, &ConfigError{
			Operation: "convert",
			Key:       key,
			Err:       fmt.Errorf("cannot convert %T to bool", value),
		}
	}
}

func (m *Manager) GetInt(key string) (int, error) {
	value, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		if err != nil {
			return 0, &ConfigError{
				Operation: "convert",
				Key:       key,
				Err:       errors.New("cannot convert string to int"),
			}
		}
		return i, nil
	default:
		return 0, &ConfigError{
			Operation: "convert",
			Key:       key,
			Err:       fmt.Errorf("cannot convert %T to int", value),
		}
	}
}

func (m *Manager) GetFloat(key string) (float64, error) {
	value, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return 0, &ConfigError{
				Operation: "convert",
				Key:       key,
				Err:       errors.New("cannot convert string to float"),
			}
		}
		return f, nil
	default:
		return 0, &ConfigError{
			Operation: "convert",
			Key:       key,
			Err:       fmt.Errorf("cannot convert %T to float", value),
		}
	}
}

func (m *Manager) GetStringSlice(key string) ([]string, error) {
	value, err := m.Get(key)
	if err != nil {
		return nil, err
	}

	if strSlice, ok := value.([]string); ok {
		return strSlice, nil
	}

	if slice, ok := value.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			switch sv := v.(type) {
			case string:
				result[i] = sv
			case []byte:
				result[i] = string(sv)
			default:
				result[i] = fmt.Sprintf("%v", v)
			}
		}
		return result, nil
	}

	if str, ok := value.(string); ok {
		return []string{str}, nil
	}

	return nil, &ConfigError{
		Operation: "convert",
		Key:       key,
		Err:       fmt.Errorf("cannot convert %T to []string", value),
	}
}
