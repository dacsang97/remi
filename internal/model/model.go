// Package model defines the application state and business logic
package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dacsang97/remi/pkg/timer"
)

// CompleteEvent represents a completed timer event
type CompleteEvent struct {
	Index int
	Name  string
	FreezeDuration time.Duration
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
	freezingEvents        map[int]time.Time
	freezeTimeRemaining   map[int]time.Duration
	logFile               *os.File
	recentLogs            []string
	mutex                 sync.Mutex
}

// NewApp creates a new application model
func NewApp(countdowns []*timer.Countdown, useSystemNotification bool) *App {
	// Create log directory if it doesn't exist
	logDir := filepath.Join(os.TempDir(), "remi-logs")
	_ = os.MkdirAll(logDir, 0755) // Ignore error
	
	// Open log file
	var logFile *os.File
	logPath := filepath.Join(logDir, "remi.log")
	logFile, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644) // Ignore error
	
	return &App{
		Countdowns:            countdowns,
		UseSystemNotification: useSystemNotification,
		freezingEvents:        make(map[int]time.Time),
		freezeTimeRemaining:   make(map[int]time.Duration),
		logFile:               logFile,
		recentLogs:            make([]string, 0, 2),
	}
}

// Tick updates all countdowns and returns any completed events
func (a *App) Tick() []CompleteEvent {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var completed []CompleteEvent

	// First, check if any freezing events have completed their freeze period
	now := time.Now()
	for index, endTime := range a.freezingEvents {
		// Update the remaining freeze time
		if now.Before(endTime) {
			a.freezeTimeRemaining[index] = endTime.Sub(now)
		} else {
			// Freeze period is over, reset the timer and start it
			if index >= 0 && index < len(a.Countdowns) {
				a.Countdowns[index].Reset()
				a.Countdowns[index].Start() // Automatically start the next round
			}
			// Remove from freezing events
			delete(a.freezingEvents, index)
			delete(a.freezeTimeRemaining, index)
		}
	}

	// Then, update all countdowns
	for i, countdown := range a.Countdowns {
		// Skip updating if this event is currently freezing
		if _, freezing := a.freezingEvents[i]; freezing {
			continue
		}

		// Update the countdown
		if complete := countdown.Update(); complete {
			// Create a complete event
			event := CompleteEvent{
				Index: i,
				Name:  countdown.Name,
				FreezeDuration: countdown.FreezeDuration,
			}
			completed = append(completed, event)

			// If this event has a freeze duration, add it to freezing events
			if countdown.FreezeDuration > 0 {
				// Calculate when the freeze period will end
				endTime := now.Add(countdown.FreezeDuration)
				a.freezingEvents[i] = endTime
				a.freezeTimeRemaining[i] = countdown.FreezeDuration
				
				// Don't reset the timer yet, it will be reset after the freeze period
				countdown.Pause()
			} else {
				// No freeze duration, reset immediately
				countdown.Reset()
			}
		}
	}

	return completed
}

// IsEventFreezing checks if an event is currently in freeze mode
func (a *App) IsEventFreezing(index int) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	_, freezing := a.freezingEvents[index]
	return freezing
}

// GetFreezeTimeRemaining returns the remaining freeze time for an event
func (a *App) GetFreezeTimeRemaining(index int) time.Duration {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.freezeTimeRemaining[index]
}

// ExecuteCommand executes a user command
func (a *App) ExecuteCommand(cmd Command) string {
	if cmd.Index < 0 || cmd.Index >= len(a.Countdowns) {
		return fmt.Sprintf("Invalid index: %d", cmd.Index)
	}

	// Check if this event is currently freezing
	if a.IsEventFreezing(cmd.Index) {
		return fmt.Sprintf("Cannot control event %d while it is freezing", cmd.Index)
	}

	countdown := a.Countdowns[cmd.Index]
	
	switch cmd.Action {
	case "start":
		countdown.Start()
		return fmt.Sprintf("Started event %d: %s", cmd.Index, countdown.Name)
	case "pause":
		countdown.Pause()
		return fmt.Sprintf("Paused event %d: %s", cmd.Index, countdown.Name)
	case "end":
		countdown.Reset()
		return fmt.Sprintf("Reset event %d: %s", cmd.Index, countdown.Name)
	default:
		return fmt.Sprintf("Unknown command: %s", cmd.Action)
	}
}

// ParseCommand parses a command string into a Command struct
func (a *App) ParseCommand(input string) (Command, error) {
	input = strings.TrimSpace(input)
	parts := strings.Fields(input)
	
	if len(parts) != 2 {
		return Command{}, fmt.Errorf("invalid command format, expected: [action] [index]")
	}
	
	action := parts[0]
	if action != "s" && action != "p" && action != "e" {
		return Command{}, fmt.Errorf("invalid action, expected: s (start), p (pause), or e (end)")
	}
	
	// Convert short form to full action name
	switch action {
	case "s":
		action = "start"
	case "p":
		action = "pause"
	case "e":
		action = "end"
	}
	
	indexStr := parts[1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return Command{}, fmt.Errorf("invalid index: %s", indexStr)
	}
	
	if index < 0 || index >= len(a.Countdowns) {
		return Command{}, fmt.Errorf("index out of range: %d", index)
	}
	
	return Command{
		Action: action,
		Index:  index,
	}, nil
}

// LogMessage logs a message to the log file and stores it for UI display
func (a *App) LogMessage(message string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// Get current timestamp
	timestamp := time.Now().Format("15:04:05")
	
	// Write to log file if available
	if a.logFile != nil {
		fullTimestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(a.logFile, "[%s] %s\n", fullTimestamp, message)
	}
	
	// Add to recent logs with timestamp (keep only the 2 most recent)
	logWithTime := fmt.Sprintf("[%s] %s", timestamp, message)
	a.recentLogs = append(a.recentLogs, logWithTime)
	if len(a.recentLogs) > 2 {
		a.recentLogs = a.recentLogs[len(a.recentLogs)-2:]
	}
}

// GetRecentLogs returns the most recent log messages for UI display
func (a *App) GetRecentLogs() []string {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// Return a copy to avoid race conditions
	logs := make([]string, len(a.recentLogs))
	copy(logs, a.recentLogs)
	return logs
}

// Close closes any resources used by the application
func (a *App) Close() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}
