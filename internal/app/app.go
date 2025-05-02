// Package app provides the main application functionality
package app

import (
	"fmt"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dacsang97/remi/internal/config"
	"github.com/dacsang97/remi/internal/model"
	"github.com/dacsang97/remi/internal/ui"
	"github.com/dacsang97/remi/pkg/timer"
)

// App represents the main application
type App struct {
	config     config.Config
	countdowns []*timer.Countdown
}

// New creates a new application instance
func New(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Parse events from configuration
	eventConfigs, err := config.ParseEvents(cfg.Events)
	if err != nil {
		return nil, fmt.Errorf("error parsing events: %w", err)
	}

	// Create countdowns from event configurations
	countdowns := make([]*timer.Countdown, 0, len(eventConfigs))
	for _, eventConfig := range eventConfigs {
		countdown := timer.NewCountdown(
			eventConfig.Name,
			eventConfig.OriginalDuration,
			eventConfig.UseNotification,
			eventConfig.IsActive,
			eventConfig.FreezeDuration,
		)
		countdowns = append(countdowns, countdown)
	}

	return &App{
		config:     cfg,
		countdowns: countdowns,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Determine if system notifications should be used
	useSystemNotification := a.config.AppConfig.UseSystemNotification && runtime.GOOS == "darwin"

	// Create application model
	app := model.NewApp(a.countdowns, useSystemNotification)

	// Create UI
	ui := ui.NewUI(app)

	// Create program
	p := tea.NewProgram(
		ui,
		tea.WithAltScreen(),
	)

	// Run program
	_, err := p.Run()
	return err
}
