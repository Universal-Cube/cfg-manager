package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// detectFileFormat determines the file format based on its extension.
// Returns the file format (Format) or an error if the format is not supported.
func detectFileFormat(filePath string) (Format, error) {
	if filePath == "" {
		return "", errors.New("empty file path")
	}

	extension := strings.ToLower(filepath.Ext(filePath))
	if extension == "" {
		return "", errors.New("no file extension found")
	}

	extension = extension[1:]

	formatMap := map[string]Format{
		"json": FormatJSON,
		"yaml": FormatYAML,
		"yml":  FormatYAML,
	}

	if format, ok := formatMap[extension]; ok {
		return format, nil
	}

	return "", fmt.Errorf("unsupported file format: %s", extension)
}

// resolvePath processes a file path by expanding environment variables and converting to absolute path.
// It supports both $VAR and ${VAR} formats for environment variables.
// Returns the resolved absolute path or an error if the path cannot be resolved.
func resolvePath(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("empty file path provided")
	}

	if strings.Contains(filePath, "$") {
		filePath = os.Expand(filePath, func(key string) string {
			return os.Getenv(key)
		})
	}

	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
		filePath = absPath
	}

	filePath = filepath.Clean(filePath)

	_, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error accessing path: %w", err)
	}

	return filePath, nil
}

// getNestedMap traverses a nested map using a dot-separated path and retrieves the target map and key.
func getNestedMap(data map[string]interface{}, path string, caseSensitive bool) (map[string]interface{}, string, error) {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return nil, "", errors.New("invalid path")
	}

	lastIndex := len(keys) - 1
	lastKey := keys[lastIndex]
	current := data

	if lastIndex == 0 {
		return current, lastKey, nil
	}

	for i := 0; i < lastIndex; i++ {
		key := keys[i]

		if !caseSensitive {
			found := false
			for k := range current {
				if strings.EqualFold(k, key) {
					key = k
					found = true
					break
				}
			}
			if !found {
				return nil, "", fmt.Errorf("key '%s' not found in path '%s'", key, path)
			}
		}

		val, exists := current[key]
		if !exists {
			return nil, "", fmt.Errorf("key '%s' not found in path '%s'", key, path)
		}

		nestedMap, ok := val.(map[string]interface{})
		if !ok {
			iFaceMap, isIFaceMap := val.(map[interface{}]interface{})
			if !isIFaceMap {
				return nil, "", fmt.Errorf("key '%s' is not a map in path '%s'", key, path)
			}

			nestedMap = make(map[string]interface{})
			for k, v := range iFaceMap {
				strKey, ok := k.(string)
				if !ok {
					return nil, "", fmt.Errorf("key '%s' is not a string in path '%s'", key, path)
				}
				nestedMap[strKey] = v
			}
		}

		current = nestedMap
	}

	if !caseSensitive {
		for k := range current {
			if strings.EqualFold(k, lastKey) {
				lastKey = k
				break
			}
		}
	}

	return current, lastKey, nil
}

// transformMapKeys recursively converts all map keys to strings within a nested structure.
// This is particularly useful when processing data loaded from YAML, which can have
// map[interface{}]interface{} types not compatible with JSON encoding.
//
// The function handles:
// - map[interface{}]interface{} -> convert to map[string]interface{}
// - map[string]interface{} -> recursively transform values
// - []interface{} -> recursively transform each element
// - other types -> returned as-is
//
// Returns the transformed structure with string keys.
func transformMapKeys(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(x))

		for k, val := range x {
			var strKey string
			switch key := k.(type) {
			case string:
				strKey = key
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				strKey = fmt.Sprintf("%d", key)
			case float32, float64:
				strKey = fmt.Sprintf("%g", key)
			case bool:
				strKey = fmt.Sprintf("%t", key)
			default:
				strKey = fmt.Sprintf("%v", key)
			}
			m[strKey] = transformMapKeys(val)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, val := range x {
			m[k] = transformMapKeys(val)
		}
		return m
	case []interface{}:
		result := make([]interface{}, len(x))
		for i, val := range x {
			result[i] = transformMapKeys(val)
		}
		return result
	default:
		return v
	}
}
