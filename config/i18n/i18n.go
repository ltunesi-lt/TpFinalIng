package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Translator struct {
	mu           sync.RWMutex
	translations map[string]map[string]string
	defaultLang  string
}

func NewTranslator(configDir string, defaultLang string) (*Translator, error) {
	t := &Translator{
		translations: make(map[string]map[string]string),
		defaultLang:  defaultLang,
	}

	files, err := filepath.Glob(filepath.Join(configDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("error al buscar archivos i18n: %w", err)
	}

	for _, file := range files {
		lang := filepath.Base(file)
		lang = lang[:len(lang)-len(filepath.Ext(lang))] // Extrae 'es', 'en', etc.

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error al leer archivo %s: %w", file, err)
		}

		var langMap map[string]string
		if err := json.Unmarshal(data, &langMap); err != nil {
			return nil, fmt.Errorf("error al parsear JSON %s: %w", file, err)
		}

		t.translations[lang] = langMap
	}

	return t, nil
}

func (t *Translator) Translate(lang string, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if langMap, exists := t.translations[lang]; exists {
		if val, ok := langMap[key]; ok {
			return val
		}
	}

	// Fallback al idioma por defecto o devolver la clave recibida
	if defaultMap, exists := t.translations[t.defaultLang]; exists {
		if val, ok := defaultMap[key]; ok {
			return val
		}
	}

	return key
}