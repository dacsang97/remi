// Package app provides the main application functionality
package app

import (
	"fmt"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dacsang97/remi/internal/config"
	"github.com/dacsang97/remi/internal/model"
	"github.com/dacsang97/remi/internal/notification"
	"github.com/dacsang97/remi/internal/ui"
	"github.com/dacsang97/remi/pkg/timer"
)

// Application represents the main application
type Application struct {
	configPath string
}

// NewApplication creates a new application instance
func NewApplication(configPath string) *Application {
	return &Application{
		configPath: configPath,
	}
}

// Run starts the application
func (a *Application) Run() error {
	// Load configuration - will try home directory first, then current directory
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Parse events from configuration
	eventConfigs, err := config.ParseEvents(cfg.Events)
	if err != nil {
		return fmt.Errorf("error parsing events: %w", err)
	}

	// Create countdowns from event configurations
	countdowns := make([]*timer.Countdown, 0, len(eventConfigs))
	for _, eventConfig := range eventConfigs {
		countdown := timer.NewCountdown(
			eventConfig.Name,
			eventConfig.OriginalDuration,
			eventConfig.UseNotification,
			eventConfig.IsActive,
		)
		countdowns = append(countdowns, countdown)
	}

	// Determine if system notifications should be used
	useSystemNotification := cfg.AppConfig.UseSystemNotification && runtime.GOOS == "darwin"

	// Create application model
	app := model.NewApp(countdowns, useSystemNotification)

	// Create notification service
	notifier := notification.NewNotifier()

	// Create UI
	ui := ui.NewUI(app, notifier)

	// Start the application with bubble tea
	p := tea.NewProgram(
		ui,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running application: %w", err)
	}

	return nil
}
