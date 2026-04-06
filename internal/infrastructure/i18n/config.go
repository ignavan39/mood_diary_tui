package i18n

type Config struct {
	LocalesPath string

	DefaultLocale Locale

	InitialLocale Locale
}

func DefaultConfig() Config {
	return Config{
		LocalesPath:   "locales",
		DefaultLocale: LocaleEN,
		InitialLocale: LocaleEN,
	}
}
