package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func detectFileFormat(filePath string) (Format, error) {
	extension := strings.ToLower(filepath.Ext(filePath))
	if len(extension) == 0 {
		return "", errors.New("no file extension found")
	}

	extension = extension[1:]

	switch extension {
	case string(FormatJSON):
		return FormatJSON, nil
	case string(FormatYAML), string(FormatYML):
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported file format: %s", extension)
	}
}

func resolvePath(filePath string) (string, error) {
	if strings.Contains(filePath, "$") {
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				name, value := parts[0], parts[1]
				filePath = strings.ReplaceAll(filePath, "$"+name, value)
				filePath = strings.ReplaceAll(filePath, "${"+name+"}", value)
			}
		}
	}

	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", err
		}
		filePath = absPath
	}

	return filePath, nil
}

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

	return nil, "", nil
}

func transformMapKeys(v interface{}) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			strKey, ok := k.(string)
			if !ok {
				strKey = fmt.Sprintf("%v", k)
			}
			m[strKey] = transformMapKeys(val)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[k] = transformMapKeys(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = transformMapKeys(val)
		}
		return v
	}

	return v
}
