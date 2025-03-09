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
