// Package freeze provides functionality to display a fullscreen overlay
package freeze

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// Service defines the interface for freeze services
type Service interface {
	Freeze(title, message string, duration time.Duration) error
	IsSupported() bool
}

// MacOSFreezeService implements freeze functionality for macOS
type MacOSFreezeService struct{}

// NewMacOSFreezeService creates a new macOS freeze service
func NewMacOSFreezeService() *MacOSFreezeService {
	return &MacOSFreezeService{}
}

// IsSupported checks if the current OS supports this freeze service
func (f *MacOSFreezeService) IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Freeze displays a fullscreen overlay for the specified duration
func (f *MacOSFreezeService) Freeze(title, message string, duration time.Duration) error {
	if !f.IsSupported() {
		return fmt.Errorf("freeze service only supports macOS")
	}

	// Calculate seconds for the countdown
	seconds := int(duration.Seconds())
	
	// Simple AppleScript dialog with countdown
	script := fmt.Sprintf(`
		tell application "System Events"
			activate
			display dialog "%s" with title "%s" buttons {"OK"} default button 1 with icon caution giving up after %d
		end tell
	`, message, title, seconds)

	// Execute the AppleScript directly without waiting
	cmd := exec.Command("osascript", "-e", script)
	
	// Log for debugging
	fmt.Printf("Executing freeze command with AppleScript dialog\n")
	
	// Start the command but don't wait for it to complete
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting AppleScript: %v\n", err)
		return fmt.Errorf("failed to start freeze command: %w", err)
	}
	
	// Run the command in a separate goroutine
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Error during AppleScript execution: %v\n", err)
		} else {
			fmt.Printf("Freeze command completed successfully\n")
		}
	}()
	
	return nil
}

// NoopFreezeService implements a no-operation freeze service for unsupported platforms
type NoopFreezeService struct{}

// NewNoopFreezeService creates a new no-operation freeze service
func NewNoopFreezeService() *NoopFreezeService {
	return &NoopFreezeService{}
}

// IsSupported always returns false for NoopFreezeService
func (f *NoopFreezeService) IsSupported() bool {
	return false
}

// Freeze is a no-op implementation that always succeeds
func (f *NoopFreezeService) Freeze(title, message string, duration time.Duration) error {
	// Do nothing, just return success
	return nil
}

// NewFreezeService creates the appropriate freeze service for the current platform
func NewFreezeService() Service {
	if runtime.GOOS == "darwin" {
		return NewMacOSFreezeService()
	}
	return NewNoopFreezeService()
}
