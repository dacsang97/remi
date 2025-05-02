// Package ui handles the terminal user interface
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dacsang97/remi/internal/model"
	"github.com/dacsang97/remi/internal/notification"
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
)

// TickMsg represents a clock tick message
type TickMsg time.Time

// CompleteMsg represents a completed timer message
type CompleteMsg struct {
	Index int
	Name  string
}

// UI represents the terminal user interface
type UI struct {
	app               *model.App
	notifier          notification.Service
	width             int
	height            int
	inputField        textinput.Model
	inputMode         bool
}

// NewUI creates a new UI instance
func NewUI(app *model.App, notifier notification.Service) *UI {
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
		notifier:          notifier,
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
			command := ui.inputField.Value()
			ui.app.LastCommandResult = ui.app.ExecuteCommand(command)

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
				return CompleteMsg{Index: event.Index, Name: event.Name}
			})
		}
		
		return ui, tea.Batch(cmds...)

	case CompleteMsg:
		// Print notification to terminal
		fmt.Printf("\n🔔 Notification: Time for '%s'!\n", msg.Name)

		// Send system notification if enabled
		if ui.app.UseSystemNotification && 
		   msg.Index >= 0 && 
		   msg.Index < len(ui.app.Countdowns) && 
		   ui.app.Countdowns[msg.Index].UseNotification {
			go func() {
				err := ui.notifier.Send("Remi Reminder", fmt.Sprintf("Time for: %s", msg.Name))
				if err != nil {
					// Log error but continue
					fmt.Printf("Error sending system notification: %v\n", err)
				}
			}()
		}
		
		return ui, nil
	}

	return ui, tea.Batch(cmds...)
}

// View renders the UI
func (ui *UI) View() string {
	if ui.width == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(" Remi - Reminder Application ") + "\n\n")

	// Display system notification status
	notificationStatus := "Off"
	if ui.app.UseSystemNotification {
		notificationStatus = "On"
	}
	s.WriteString(fmt.Sprintf("macOS System Notifications: %s\n\n", notificationStatus))

	// Calculate optimal container width
	containerWidth := ui.width - 4
	if containerWidth < 40 {
		containerWidth = 40 // Minimum width
	}

	// Create a single container for all events with a simple style
	var events []string
	
	// Render each countdown
	for i, countdown := range ui.app.Countdowns {
		hours := int(countdown.RemainingTime.Hours())
		minutes := int(countdown.RemainingTime.Minutes()) % 60
		seconds := int(countdown.RemainingTime.Seconds()) % 60

		timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		percentRemaining := countdown.GetPercentRemaining()

		// Show notification icon if enabled
		notifyIcon := " "
		if countdown.UseNotification && ui.app.UseSystemNotification {
			notifyIcon = "🔔 "
		}

		// Show activity status
		statusText := ""
		if countdown.IsActive {
			statusText = activeStyle.Render("▶ Active")
		} else {
			statusText = pausedStyle.Render("⏸ Paused")
		}

		// Create title line with event name and time
		indexStr := fmt.Sprintf("[%d] ", i)
		
		// Calculate available width for name to prevent UI breaking
		reservedSpace := 30 // Approximate space for other elements
		availableWidth := containerWidth - reservedSpace
		
		// Truncate name if it's too long
		displayName := countdown.Name
		if len(displayName)*2 > availableWidth { // Rough estimate for character width
			maxLen := availableWidth / 2
			if maxLen > 3 {
				displayName = displayName[:maxLen-3] + "..."
			}
		}
		
		// Create a single line for each event
		eventLine := lipgloss.JoinHorizontal(
			lipgloss.Left,
			indexStr,
			notifyIcon,
			nameStyle.Render(displayName),
			"  ",
			timeStyle.Render(timeStr),
			"  ",
			fmt.Sprintf("(%d%%)", percentRemaining),
			"  ",
			statusText,
		)
		
		// Add event to the list
		events = append(events, eventLine)
	}
	
	// Style for the event box
	eventBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(1).
		Width(containerWidth)
		
	// Render each event in its own box
	for _, event := range events {
		s.WriteString(eventBoxStyle.Render(event) + "\n\n")
	}

	// Display last command result if available
	if ui.app.LastCommandResult != "" {
		resultStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Padding(0, 1)
		s.WriteString(resultStyle.Render("Command result: "+ui.app.LastCommandResult) + "\n\n")
	}

	// Display input field with style
	inputWidth := ui.width - 8
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	} else if inputWidth > 60 {
		inputWidth = 60 // Maximum width
	}
	
	responsiveInputStyle := inputStyle.Copy().Width(inputWidth)
	s.WriteString(responsiveInputStyle.Render(ui.inputField.View()) + "\n\n")
	s.WriteString("Commands: s = start, p = pause, e = end (reset)\n")
	s.WriteString("Example: 's 0' to start event with index 0\n\n")
	s.WriteString("Press ESC or Ctrl+C to exit\n")

	return s.String()
}

// tickEverySecond creates a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
