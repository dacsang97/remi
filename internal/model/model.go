// Package model defines the application state and business logic
package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dacsang97/remi/pkg/timer"
)

// CompleteEvent represents a completed timer event
type CompleteEvent struct {
	Index int
	Name  string
}

// Command represents a user command with parsed arguments
type Command struct {
	Action string
	Index  int
}

// App represents the application state and business logic
type App struct {
	Countdowns            []*timer.Countdown
	UseSystemNotification bool
	LastCommandResult     string
}

// NewApp creates a new application model
func NewApp(countdowns []*timer.Countdown, useSystemNotification bool) *App {
	return &App{
		Countdowns:            countdowns,
		UseSystemNotification: useSystemNotification,
	}
}

// Tick updates all active countdowns and returns completed events
func (a *App) Tick() []CompleteEvent {
	var completed []CompleteEvent

	for i, countdown := range a.Countdowns {
		if countdown.Tick(time.Second) {
			completed = append(completed, CompleteEvent{
				Index: i,
				Name:  countdown.Name,
			})
			// Reset the timer after completion
			countdown.Reset()
		}
	}

	return completed
}

// ExecuteCommand processes a user command and updates the model
func (a *App) ExecuteCommand(commandStr string) string {
	cmd, err := parseCommand(commandStr)
	if err != nil {
		return err.Error()
	}

	// Check if index is valid
	if cmd.Index < 0 || cmd.Index >= len(a.Countdowns) {
		return fmt.Sprintf("Index out of range: %d", cmd.Index)
	}

	// Process command
	switch cmd.Action {
	case "s": // Start
		a.Countdowns[cmd.Index].Start()
		return fmt.Sprintf("Started '%s'", a.Countdowns[cmd.Index].Name)

	case "p": // Pause
		a.Countdowns[cmd.Index].Pause()
		return fmt.Sprintf("Paused '%s'", a.Countdowns[cmd.Index].Name)

	case "e": // End/Reset
		a.Countdowns[cmd.Index].End()
		return fmt.Sprintf("Reset '%s'", a.Countdowns[cmd.Index].Name)

	default:
		return fmt.Sprintf("Unknown command: %s", cmd.Action)
	}
}

// parseCommand parses a command string into a structured Command
func parseCommand(commandStr string) (Command, error) {
	parts := strings.Fields(commandStr)
	if len(parts) < 2 {
		return Command{}, fmt.Errorf("Invalid command. Use: s/p/e + index")
	}

	action := parts[0]
	indexStr := parts[1]

	// Parse index
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return Command{}, fmt.Errorf("Invalid index: %s", indexStr)
	}

	return Command{
		Action: action,
		Index:  index,
	}, nil
}
