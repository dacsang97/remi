// Package ui handles the terminal user interface
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dacsang97/remi/internal/freeze"
	"github.com/dacsang97/remi/internal/model"
)

// Styles for UI components
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		PaddingLeft(1).
		PaddingRight(1)

	nameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		PaddingLeft(1)

	timeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87")).
		Bold(true).
		PaddingLeft(1)

	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(0, 1)

	activeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).
		Bold(true)

	pausedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87")).
		Bold(true)
		
	freezeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true)
		
	logStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFA500")).
		Padding(0, 1)
)

// TickMsg represents a clock tick message
type TickMsg time.Time

// CompleteMsg represents a completed timer message
type CompleteMsg struct {
	Index int
	Name  string
	FreezeDuration time.Duration
}

// UI represents the terminal user interface
type UI struct {
	app               *model.App
	freezer           freeze.Service
	width             int
	height            int
	inputField        textinput.Model
	inputMode         bool
}

// NewUI creates a new UI instance
func NewUI(app *model.App) *UI {
	// Initialize input field
	ti := textinput.New()
	ti.Placeholder = "Enter command (s/p/e + index) e.g.: s 1"
	ti.Focus()
	ti.CharLimit = 30
	ti.Width = 30
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#25A065"))

	return &UI{
		app:               app,
		freezer:           freeze.NewFreezeService(),
		inputField:        ti,
		inputMode:         true,
	}
}

// Init initializes the UI
func (ui *UI) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickEverySecond(),
	)
}

// Update handles UI events and updates
func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return ui, tea.Quit
		}

		// Handle keyboard input for command input
		var cmd tea.Cmd
		ui.inputField, cmd = ui.inputField.Update(msg)

		// Check for Enter key press
		if msg.Type == tea.KeyEnter {
			// Parse command from input field
			commandStr := ui.inputField.Value()
			cmd, err := ui.app.ParseCommand(commandStr)
			if err != nil {
				ui.app.LastCommandResult = err.Error()
			} else {
				ui.app.LastCommandResult = ui.app.ExecuteCommand(cmd)
			}

			// Clear input field after executing command
			ui.inputField.SetValue("")
		}

		cmds = append(cmds, cmd)
		return ui, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		ui.width = msg.Width
		ui.height = msg.Height
		return ui, nil

	case TickMsg:
		cmds = append(cmds, tickEverySecond())

		// Process timer ticks and collect completed events
		completedEvents := ui.app.Tick()
		
		// Create commands for completed events
		for _, event := range completedEvents {
			cmds = append(cmds, func() tea.Msg {
				return CompleteMsg{Index: event.Index, Name: event.Name, FreezeDuration: event.FreezeDuration}
			})
		}
		
		return ui, tea.Batch(cmds...)

	case CompleteMsg:
		// Log notification to file instead of terminal
		ui.app.LogMessage(fmt.Sprintf("Notification: Time for '%s'!", msg.Name))

		// Activate freeze mode if enabled for this event
		if msg.FreezeDuration > 0 {
			go func() {
				ui.app.LogMessage(fmt.Sprintf("Activating freeze mode for %s with duration %s", 
					msg.Name, msg.FreezeDuration.String()))
				
				err := ui.freezer.Freeze("Remi Reminder", fmt.Sprintf("Time for: %s\nThis screen will be locked for %s", 
					msg.Name, msg.FreezeDuration.String()), msg.FreezeDuration)
				if err != nil {
					// Log error but continue
					ui.app.LogMessage(fmt.Sprintf("Error activating freeze mode: %v", err))
				}
			}()
		} else {
			// If no freeze duration, just show a notification dialog
			go func() {
				err := ui.freezer.Freeze("Remi Reminder", fmt.Sprintf("Time for: %s", msg.Name), 5*time.Second)
				if err != nil {
					ui.app.LogMessage(fmt.Sprintf("Error showing notification dialog: %v", err))
				}
			}()
		}
		
		return ui, nil
	}

	return ui, tea.Batch(cmds...)
}

// View renders the UI
func (ui *UI) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Remi - Reminder Application ") + "\n\n")

	// Calculate optimal container width
	containerWidth := ui.width - 4
	if containerWidth < 40 {
		containerWidth = 40
	}

	// Display each event
	for i, countdown := range ui.app.Countdowns {
		// Create a container for this event
		container := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Width(containerWidth).
			Align(lipgloss.Left).
			MarginBottom(1)

		// Format the index
		indexStr := fmt.Sprintf("[%d]", i)

		// Format the name (truncate if too long)
		maxNameLength := containerWidth - 40 // Reserve space for other elements
		if maxNameLength < 10 {
			maxNameLength = 10
		}
		displayName := countdown.Name
		if len(displayName) > maxNameLength {
			displayName = displayName[:maxNameLength-3] + "..."
		}

		// Format the time
		remainingTime := countdown.RemainingTime
		hours := int(remainingTime.Hours())
		minutes := int(remainingTime.Minutes()) % 60
		seconds := int(remainingTime.Seconds()) % 60
		timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		percentRemaining := countdown.GetPercentRemaining()

		// Show activity status
		statusText := ""
		
		// Check if the timer is in freeze mode
		if ui.app.IsEventFreezing(i) {
			// Show freeze countdown
			freezeRemaining := ui.app.GetFreezeTimeRemaining(i)
			minutes := int(freezeRemaining.Minutes())
			seconds := int(freezeRemaining.Seconds()) % 60
			freezeTimeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)
			statusText = freezeStyle.Render(fmt.Sprintf("⏸ Freezing (%s)", freezeTimeStr))
		} else if countdown.IsActive {
			statusText = activeStyle.Render("▶ Active")
		} else {
			statusText = pausedStyle.Render("⏸ Paused")
		}

		// Format the percentage
		percentStr := fmt.Sprintf("(%d%%)", percentRemaining)

		// Combine all elements into a single line
		eventLine := lipgloss.JoinHorizontal(
			lipgloss.Left,
			indexStr,
			"  ",
			nameStyle.Render(displayName),
			"   ",
			timeStyle.Render(timeStr),
			"  ",
			percentStr,
			"  ",
			statusText,
		)

		// Add the event to the container and append to output
		s.WriteString(container.Render("  " + eventLine + "  ") + "\n")
	}

	// Input field
	inputContainer := inputStyle.Width(containerWidth / 2)
	s.WriteString(inputContainer.Render(ui.inputField.View()) + "\n\n")

	// Help text
	s.WriteString("Commands: s = start, p = pause, e = end (reset)\n")
	s.WriteString("Example: 's 0' to start event with index 0\n\n")
	
	// Last command result
	if ui.app.LastCommandResult != "" {
		s.WriteString(ui.app.LastCommandResult + "\n\n")
	}

	s.WriteString("Press ESC or Ctrl+C to exit\n\n")

	// Display recent logs after help text
	logs := ui.app.GetRecentLogs()
	if len(logs) > 0 {
		logContainer := logStyle.Width(containerWidth)
		logContent := "📋 Recent Activity:\n"
		for _, log := range logs {
			logContent += "• " + log + "\n"
		}
		s.WriteString(logContainer.Render(logContent))
	}

	return s.String()
}

// tickEverySecond creates a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
