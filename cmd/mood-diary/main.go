package main

import (
	"context"
	"fmt"
	"mood-diary/internal/application/usecase"
	"mood-diary/internal/infrastructure/database"
	persistence "mood-diary/internal/infrastructure/persistence/repository"
	"mood-diary/internal/presentation/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

	repo := persistence.NewSQLiteMoodRepository(db.DB())

	service := usecase.NewMoodService(repo)

	ctx := context.Background()

	model := tui.NewModel(ctx, service)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
