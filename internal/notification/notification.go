// Package notification handles system notifications
package notification

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Service defines the interface for notification services
type Service interface {
	Send(title, message string) error
	IsSupported() bool
}

// MacOSNotifier implements notification for macOS
type MacOSNotifier struct{}

// NewMacOSNotifier creates a new macOS notification service
func NewMacOSNotifier() *MacOSNotifier {
	return &MacOSNotifier{}
}

// IsSupported checks if the current OS supports this notification service
func (n *MacOSNotifier) IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Send sends a notification using macOS notification system
func (n *MacOSNotifier) Send(title, message string) error {
	if !n.IsSupported() {
		return fmt.Errorf("notification service only supports macOS")
	}

	// Try terminal-notifier first as it's more reliable
	if _, err := exec.LookPath("terminal-notifier"); err == nil {
		cmd := exec.Command("terminal-notifier",
			"-title", title,
			"-message", message,
			"-sound", "default")
		return cmd.Run()
	}

	// Fall back to osascript if terminal-notifier isn't available
	script := fmt.Sprintf(`
		tell application "System Events"
			display notification "%s" with title "%s" sound name "Glass"
		end tell`, message, title)

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// NoopNotifier implements a no-operation notifier for unsupported platforms
type NoopNotifier struct{}

// NewNoopNotifier creates a new no-operation notification service
func NewNoopNotifier() *NoopNotifier {
	return &NoopNotifier{}
}

// IsSupported always returns false for NoopNotifier
func (n *NoopNotifier) IsSupported() bool {
	return false
}

// Send is a no-op implementation that always succeeds
func (n *NoopNotifier) Send(title, message string) error {
	// Do nothing, just return success
	return nil
}

// NewNotifier creates the appropriate notifier for the current platform
func NewNotifier() Service {
	if runtime.GOOS == "darwin" {
		return NewMacOSNotifier()
	}
	return NewNoopNotifier()
}
