package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	translations map[string]map[string]string
	currentLang  string
	mu           sync.RWMutex
)

// setLanguageInternal sets the language without locking (caller must hold lock)
func setLanguageInternal(langCode string) error {
	slog.Info("Setting language", "language", langCode)

	// Load translation file
	filePath := fmt.Sprintf("locales/%s.json", langCode)
	slog.Debug("Reading locale file", "file", filePath)
	data, err := localesFS.ReadFile(filePath)
	if err != nil {
		slog.Error("Failed to read locale file", "file", filePath, "error", err)
		return fmt.Errorf("failed to load locale file for %s: %w", langCode, err)
	}
	slog.Debug("Locale file read successfully", "file", filePath, "size_bytes", len(data))

	slog.Debug("Unmarshaling locale JSON")
	var langTranslations map[string]string
	if err := json.Unmarshal(data, &langTranslations); err != nil {
		slog.Error("Failed to unmarshal locale JSON", "file", filePath, "error", err)
		return fmt.Errorf("failed to parse locale file for %s: %w", langCode, err)
	}
	slog.Info("Locale loaded successfully", "language", langCode, "translations_count", len(langTranslations))

	translations[langCode] = langTranslations
	currentLang = langCode
	slog.Debug("Language set successfully", "language", langCode)
	return nil
}

// Init initializes the i18n system with the given language code
func Init(langCode string) error {
	slog.Info("Initializing i18n", "language", langCode)
	mu.Lock()
	defer mu.Unlock()

	translations = make(map[string]map[string]string)
	err := setLanguageInternal(langCode)
	if err != nil {
		slog.Error("Failed to initialize i18n", "language", langCode, "error", err)
	} else {
		slog.Info("i18n initialized successfully", "language", langCode)
	}
	return err
}

// SetLanguage sets the current language
func SetLanguage(langCode string) error {
	slog.Info("Setting language", "language", langCode)
	mu.Lock()
	defer mu.Unlock()

	return setLanguageInternal(langCode)
}

// T translates a key with optional arguments
func T(key string, args ...interface{}) string {
	mu.RLock()
	defer mu.RUnlock()

	langTranslations, ok := translations[currentLang]
	if !ok {
		return key
	}

	translation, ok := langTranslations[key]
	if !ok {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(translation, args...)
	}

	return translation
}

// GetCurrentLanguage returns the current language code
func GetCurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// GetLanguageName returns the human-readable name for a language code
func GetLanguageName(langCode string) string {
	names := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
	}
	if name, ok := names[langCode]; ok {
		return name
	}
	return langCode
}

// GetSupportedLanguages returns a list of supported language codes
func GetSupportedLanguages() []string {
	return []string{"en", "es", "fr", "de"}
}
