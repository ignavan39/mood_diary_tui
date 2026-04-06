package i18n

import (
	"fmt"
	"sync"
)

type Translator interface {
	T(key string, args ...any) string

	SetLocale(locale Locale) error

	Locale() Locale

	SupportedLocales() []Locale
}

type translator struct {
	mu            sync.RWMutex
	basePath      string
	currentLocale Locale

	cache map[Locale]TranslationMap

	fallback TranslationMap
}

func NewTranslator(basePath string, fallback Locale) (Translator, error) {
	t := &translator{
		basePath:      basePath,
		currentLocale: fallback,
		cache:         make(map[Locale]TranslationMap),
		fallback:      nil,
	}

	fallbackMap, err := LoadTranslations(fallback, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load fallback locale %s: %w", fallback, err)
	}
	t.fallback = fallbackMap
	t.cache[fallback] = fallbackMap

	for _, loc := range SupportedLocales() {
		if loc == fallback {
			continue
		}
		if m, err := LoadTranslations(loc, basePath); err == nil {
			t.cache[loc] = m
		}

	}

	return t, nil
}

func (t *translator) T(key string, args ...any) string {
	t.mu.RLock()
	currentMap := t.cache[t.currentLocale]
	fallbackMap := t.fallback
	t.mu.RUnlock()

	if template, ok := currentMap[key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(template, args...)
		}
		return template
	}

	if fallbackMap != nil {
		if template, ok := fallbackMap[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(template, args...)
			}
			return template
		}
	}

	return key
}

func (t *translator) SetLocale(locale Locale) error {
	if !locale.IsValid() {
		return fmt.Errorf("unsupported locale: %s", locale)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.cache[locale]; exists {
		t.currentLocale = locale
		return nil
	}

	m, err := LoadTranslations(locale, t.basePath)
	if err != nil {
		return fmt.Errorf("failed to load locale %s: %w", locale, err)
	}
	t.cache[locale] = m
	t.currentLocale = locale
	return nil
}

func (t *translator) Locale() Locale {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLocale
}

func (t *translator) SupportedLocales() []Locale {
	return SupportedLocales()
}
