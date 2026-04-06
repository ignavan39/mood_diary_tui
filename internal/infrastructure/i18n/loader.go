package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type TranslationMap map[string]string

func LoadTranslations(locale Locale, basePath string) (TranslationMap, error) {

	if !locale.IsValid() {
		return nil, fmt.Errorf("unsupported locale: %s", locale)
	}

	cleanBase := filepath.Clean(basePath)
	if !filepath.IsAbs(cleanBase) {
		var err error
		cleanBase, err = filepath.Abs(cleanBase)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve base path: %w", err)
		}
	}

	filename := locale.ToFilename()
	fullPath := filepath.Join(cleanBase, filename)

	relPath, err := filepath.Rel(cleanBase, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return nil, fmt.Errorf("path traversal attempt detected: %s", fullPath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("locale file not found: %s", fullPath)
		}
		return nil, fmt.Errorf("failed to read locale file: %w", err)
	}

	var raw map[string]interface{}
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	result := make(TranslationMap)
	flattenTOML(raw, "", result)

	return result, nil
}

func flattenTOML(input map[string]any, prefix string, output TranslationMap) {
	for key, value := range input {

		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:

			if strings.TrimSpace(v) != "" {
				output[fullKey] = v
			}
		case map[string]any:

			flattenTOML(v, fullKey, output)
		case map[any]any:

			flattened := make(map[string]any)
			for k, val := range v {
				if ks, ok := k.(string); ok {
					flattened[ks] = val
				}
			}
			flattenTOML(flattened, fullKey, output)
		}
	}
}

func MergeWithFallback(primary, fallback TranslationMap) TranslationMap {
	result := make(TranslationMap, len(primary)+len(fallback))

	for k, v := range fallback {
		result[k] = v
	}

	for k, v := range primary {
		result[k] = v
	}

	return result
}
