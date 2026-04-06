package entity

import (
	"errors"
	"time"
)

type SettingsKey string

const (
	SettingsKeyLanguage SettingsKey = "language"
)

type UserSettings struct {
	Key       SettingsKey
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *UserSettings) Validate() error {
	switch s.Key {
	case SettingsKeyLanguage:
		if len(s.Value) < 2 || len(s.Value) > 10 {
			return ErrInvalidLanguageCode
		}
		return nil
	default:
		return ErrUnknownSettingsKey
	}
}

var (
	ErrInvalidLanguageCode = errors.New("invalid language code format")
	ErrUnknownSettingsKey  = errors.New("unknown settings key")
	ErrSettingsNotFound    = errors.New("settings not found")
)
