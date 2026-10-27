package i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lt-tunesi/TpFinalIng/config/i18n"
)

func TestTranslator_Translate(t *testing.T) {
	// Crear un directorio temporal con archivos JSON de prueba
	tempDir := t.TempDir()

	esContent := []byte(`{"welcome": "Bienvenido", "bye": "Adiós"}`)
	enContent := []byte(`{"welcome": "Welcome"}`)

	if err := os.WriteFile(filepath.Join(tempDir, "es.json"), esContent, 0644); err != nil {
		t.Fatalf("error creando es.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "en.json"), enContent, 0644); err != nil {
		t.Fatalf("error creando en.json: %v", err)
	}

	// Inicializar traductor
	translator, err := i18n.NewTranslator(tempDir, "es")
	if err != nil {
		t.Fatalf("error inicializando traductor: %v", err)
	}

	// 1. Traducción existente en español
	if got := translator.Translate("es", "welcome"); got != "Bienvenido" {
		t.Errorf("esperado 'Bienvenido', obtenido '%s'", got)
	}

	// 2. Traducción existente en inglés
	if got := translator.Translate("en", "welcome"); got != "Welcome" {
		t.Errorf("esperado 'Welcome', obtenido '%s'", got)
	}

	// 3. Fallback al idioma por defecto ('es') si falta la clave en 'en'
	if got := translator.Translate("en", "bye"); got != "Adiós" {
		t.Errorf("esperado 'Adiós' (fallback), obtenido '%s'", got)
	}

	// 4. Devolver la clave original si no existe en ningún idioma
	if got := translator.Translate("en", "non_existent_key"); got != "non_existent_key" {
		t.Errorf("esperado 'non_existent_key', obtenido '%s'", got)
	}
}