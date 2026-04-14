package i18n

import (
	"fmt"
	"strings"
)

type Locale string

const (
	LocaleEN Locale = "en"
	LocaleRU Locale = "ru"
	LocaleJA Locale = "ja"
)

func SupportedLocales() []Locale {
	return []Locale{LocaleEN, LocaleRU, LocaleJA}
}

func (l Locale) IsValid() bool {
	for _, supported := range SupportedLocales() {
		if l == supported {
			return true
		}
	}
	return false
}

func NormalizeLocale(input string) Locale {

	code := strings.ToLower(strings.TrimSpace(input))
	if len(code) >= 2 {
		code = code[:2]
	}

	locale := Locale(code)
	if locale.IsValid() {
		return locale
	}
	return LocaleEN
}

func (l Locale) ToFilename() string {
	return fmt.Sprintf("%s.toml", string(l))
}
