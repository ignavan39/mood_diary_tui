package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/infrastructure/database"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/infrastructure/persistence"
	"github.com/ignavan39/mood-diary/internal/presentation/tui"
)

func main() {

	dbPath, err := database.GetDefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting database path: %v\n", err)
		os.Exit(1)
	}

	db, err := database.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	moodRepo := persistence.NewSQLiteMoodRepository(db.DB())
	settingsRepo := persistence.NewSQLiteSettingsRepository(db.DB())

	moodService := usecase.NewMoodService(moodRepo)

	ctx := context.Background()
	translator, err := setupI18n(settingsRepo, ctx)
	if err != nil {
		log.Printf("i18n initialization warning: %v", err)

		translator, _ = i18n.NewTranslator("locales", i18n.LocaleEN)
	}

	model := tui.NewModel(ctx, moodService, translator, settingsRepo)

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func setupI18n(
	settingsRepo repository.SettingsRepository,
	ctx context.Context,
) (i18n.Translator, error) {

	basePath := determineLocalesPath()

	cfg := i18n.Config{
		LocalesPath:   basePath,
		DefaultLocale: i18n.LocaleEN,
		InitialLocale: i18n.LocaleEN,
	}

	translator, err := i18n.NewTranslator(cfg.LocalesPath, cfg.DefaultLocale)
	if err != nil {
		return nil, fmt.Errorf("failed to create translator: %w", err)
	}

	if setting, err := settingsRepo.Get(ctx, entity.SettingsKeyLanguage); err == nil {
		userLocale := i18n.NormalizeLocale(setting.Value)
		if userLocale != cfg.DefaultLocale {
			if err := translator.SetLocale(userLocale); err != nil {
				log.Printf("Could not set user locale %s: %v (fallback to %s)",
					userLocale, err, cfg.DefaultLocale)
			} else {
				log.Printf("✓ User locale set to: %s", userLocale)
			}
		}
	} else {

		defaultSetting := &entity.UserSettings{
			Key:   entity.SettingsKeyLanguage,
			Value: string(cfg.DefaultLocale),
		}
		if upsertErr := settingsRepo.Upsert(ctx, defaultSetting); upsertErr != nil {
			log.Printf("Could not create default language setting: %v", upsertErr)
		}
	}

	return translator, nil
}

func determineLocalesPath() string {

	if envPath := os.Getenv("MOOD_DIARY_LOCALES"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
		log.Printf("MOOD_DIARY_LOCALES=%s not found, trying defaults", envPath)
	}

	if _, err := os.Stat("locales"); err == nil {
		return "locales"
	}

	if _, err := os.Stat("../locales"); err == nil {
		return "../locales"
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, "locales")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "locales"
}
