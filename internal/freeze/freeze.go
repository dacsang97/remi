// Package freeze provides functionality to display a fullscreen overlay
package freeze

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Logger provides a simple logging interface
type Logger interface {
	Printf(format string, v ...interface{})
}

// FileLogger implements a simple file logger
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a new file logger
func NewFileLogger(filename string) (*FileLogger, error) {
	// Create log directory if it doesn't exist
	logDir := filepath.Join(os.TempDir(), "remi-logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	
	// Open log file
	logPath := filepath.Join(logDir, filename)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	
	return &FileLogger{file: file}, nil
}

// Printf logs a formatted message to the file
func (l *FileLogger) Printf(format string, v ...interface{}) {
	if l.file != nil {
		fmt.Fprintf(l.file, format+"\n", v...)
	}
}

// Close closes the log file
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// NoopLogger implements a logger that does nothing
type NoopLogger struct{}

// Printf does nothing for NoopLogger
func (l *NoopLogger) Printf(format string, v ...interface{}) {}

// Service defines the interface for freeze services
type Service interface {
	Freeze(title, message string, duration time.Duration) error
	IsSupported() bool
	Close() error
}

// MacOSFreezeService implements freeze functionality for macOS
type MacOSFreezeService struct {
	logger Logger
}

// NewMacOSFreezeService creates a new macOS freeze service
func NewMacOSFreezeService() *MacOSFreezeService {
	logger, err := NewFileLogger("freeze.log")
	if err != nil {
		// Fall back to noop logger if file logger fails
		return &MacOSFreezeService{logger: &NoopLogger{}}
	}
	return &MacOSFreezeService{logger: logger}
}

// IsSupported checks if the current OS supports this freeze service
func (f *MacOSFreezeService) IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Close closes any resources used by the service
func (f *MacOSFreezeService) Close() error {
	if closer, ok := f.logger.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
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
	f.logger.Printf("Executing freeze command with AppleScript dialog")
	
	// Start the command but don't wait for it to complete
	if err := cmd.Start(); err != nil {
		f.logger.Printf("Error starting AppleScript: %v", err)
		return fmt.Errorf("failed to start freeze command: %w", err)
	}
	
	// Run the command in a separate goroutine
	go func() {
		if err := cmd.Wait(); err != nil {
			f.logger.Printf("Error during AppleScript execution: %v", err)
		} else {
			f.logger.Printf("Freeze command completed successfully")
		}
	}()
	
	return nil
}

// NoopFreezeService implements a no-operation freeze service for unsupported platforms
type NoopFreezeService struct {
	logger Logger
}

// NewNoopFreezeService creates a new no-operation freeze service
func NewNoopFreezeService() *NoopFreezeService {
	return &NoopFreezeService{logger: &NoopLogger{}}
}

// IsSupported always returns false for NoopFreezeService
func (f *NoopFreezeService) IsSupported() bool {
	return false
}

// Close is a no-op for NoopFreezeService
func (f *NoopFreezeService) Close() error {
	return nil
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
