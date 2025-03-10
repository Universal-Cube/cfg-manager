package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func (m *Manager) Set(key string, value interface{}) error {
	if key == "" {
		return &ConfigError{
			Operation: "set",
			Key:       key,
			Err:       errors.New("key cannot be empty"),
		}
	}

	keys := strings.Split(key, ".")
	lastIndex := len(keys) - 1
	lastKey := keys[lastIndex]

	if lastIndex == 0 {
		m.data[lastKey] = value
		return nil
	}

	current := m.data
	for i := 0; i < lastIndex; i++ {
		k := keys[i]
		v, exists := current[k]
		if !exists {
			newMap := make(map[string]interface{})
			current[k] = newMap
			current = newMap
			continue
		}

		nextMap, ok := v.(map[string]interface{})
		if !ok {
			iFaceMap, isIFaceMap := v.(map[interface{}]interface{})
			if !isIFaceMap {
				newMap := make(map[string]interface{})
				current[k] = newMap
				current = newMap
				continue
			}

			nextMap = make(map[string]interface{})
			for ik, iv := range iFaceMap {
				strKey, ok := ik.(string)
				if !ok {
					strKey = fmt.Sprintf("%v", ik)
				}
				nextMap[strKey] = iv
			}
			current[k] = nextMap
		}

		current = nextMap
	}

	current[lastKey] = value
	return nil
}

func (m *Manager) Has(key string) bool {
	_, err := m.Get(key)
	return err == nil
}

func (m *Manager) Delete(key string) error {
	if key == "" {
		return &ConfigError{
			Operation: "delete",
			Key:       key,
			Err:       errors.New("empty key"),
		}
	}

	parentMap, lastKey, err := getNestedMap(m.data, key, m.caseSensitive)
	if err != nil {
		return &ConfigError{
			Operation: "delete",
			Key:       key,
			Err:       err,
		}
	}

	if !m.caseSensitive {
		for k := range parentMap {
			if strings.EqualFold(k, lastKey) {
				lastKey = k
				break
			}
		}
	}

	if _, exists := parentMap[lastKey]; !exists {
		return &ConfigError{
			Operation: "delete",
			Key:       key,
			Err:       errors.New("key not found"),
		}
	}

	delete(parentMap, lastKey)
	return nil
}

func (m *Manager) Save() error {
	if m.filePath == "" {
		return &ConfigError{
			Operation: "save",
			Err:       errors.New("file path not set"),
		}
	}

	return m.SaveToFile(m.filePath, m.fileFormat)
}

func (m *Manager) SaveToFile(path string, format Format) error {
	resolvedPath, err := resolvePath(path)
	if err != nil {
		return &ConfigError{
			Operation: "resolve path",
			Err:       err,
		}
	}

	var content []byte
	switch format {
	case FormatJSON:
		content, err = json.MarshalIndent(m.data, "", "  ")
	case FormatYAML, FormatYML:
		content, err = yaml.Marshal(m.data)
	default:
		err = fmt.Errorf("unsupported file format: %s", format)
	}

	if err != nil {
		return &ConfigError{
			Operation: "marshal",
			Err:       err,
		}
	}

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ConfigError{
			Operation: "create directory",
			Err:       err,
		}
	}

	if err := os.WriteFile(resolvedPath, content, 0644); err != nil {
		return &ConfigError{
			Operation: "write file",
			Err:       err,
		}
	}

	m.filePath = resolvedPath
	m.fileFormat = format

	return nil
}

func (m *Manager) Data() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m.data {
		result[k] = v
	}

	return result
}

func (m *Manager) WithFilePath(path string) *Manager {
	m.filePath = path
	return m
}

func (m *Manager) WithFormat(format Format) *Manager {
	m.fileFormat = format
	return m
}

func (m *Manager) Clear() {
	m.data = make(map[string]interface{})
}

func (m *Manager) Merge(other *Manager) {
	m.MergeMap(other.data)
}

func (m *Manager) MergeMap(data map[string]interface{}) {
	for k, v := range data {
		if existing, ok := m.data[k]; ok {
			if existingMap, isMap := existing.(map[string]interface{}); isMap {
				if newMap, isMap := v.(map[string]interface{}); isMap {
					for nk, nv := range newMap {
						existingMap[nk] = nv
					}
					continue
				}
			}
		}

		m.data[k] = v
	}
}

func (m *Manager) ThreadSafe() *ThreadSafeManager {
	return &ThreadSafeManager{
		manager: m,
		mu:      &sync.RWMutex{},
	}
}

func (t *ThreadSafeManager) Get(key string) (interface{}, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.Get(key)
}

func (t *ThreadSafeManager) GetString(key string) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.GetString(key)
}

func (t *ThreadSafeManager) GetBool(key string) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.GetBool(key)
}

func (t *ThreadSafeManager) GetInt(key string) (int, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.GetInt(key)
}

func (t *ThreadSafeManager) GetFloat(key string) (float64, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.GetFloat(key)
}

func (t *ThreadSafeManager) GetStringSlice(key string) ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.GetStringSlice(key)
}

func (t *ThreadSafeManager) Set(key string, value interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.manager.Set(key, value)
}

func (t *ThreadSafeManager) Has(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.Has(key)
}

func (t *ThreadSafeManager) Delete(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.manager.Delete(key)
}

func (t *ThreadSafeManager) Save() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.Save()
}

func (t *ThreadSafeManager) SaveToFile(path string, format Format) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.SaveToFile(path, format)
}

func (t *ThreadSafeManager) Data() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.manager.Data()
}

func (t *ThreadSafeManager) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.manager.Clear()
}

func (t *ThreadSafeManager) Merge(other *Manager) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.manager.Merge(other)
}

func (t *ThreadSafeManager) MergeMap(data map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.manager.MergeMap(data)
}
