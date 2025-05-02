// Package timer provides countdown timer functionality
package timer

import (
	"fmt"
	"time"
)

// Countdown represents a countdown timer for an event
type Countdown struct {
	Name             string
	OriginalDuration time.Duration
	RemainingTime    time.Duration
	UseNotification  bool
	IsActive         bool
	FreezeDuration   time.Duration
}

// NewCountdown creates a new countdown timer
func NewCountdown(name string, duration time.Duration, useNotification bool, autoStart bool, freezeDuration time.Duration) *Countdown {
	return &Countdown{
		Name:             name,
		OriginalDuration: duration,
		RemainingTime:    duration,
		UseNotification:  useNotification,
		IsActive:         autoStart,
		FreezeDuration:   freezeDuration,
	}
}

// Tick decrements the timer by the specified duration if active
// Returns true if the timer has completed this tick
func (c *Countdown) Tick(d time.Duration) bool {
	if !c.IsActive || c.RemainingTime <= 0 {
		return false
	}

	c.RemainingTime -= d
	return c.RemainingTime <= 0
}

// Update decrements the timer by one second if active
// Returns true if the timer has completed this tick
func (c *Countdown) Update() bool {
	return c.Tick(time.Second)
}

// Reset resets the timer to its original duration
func (c *Countdown) Reset() {
	c.RemainingTime = c.OriginalDuration
}

// Start activates the timer
func (c *Countdown) Start() {
	c.IsActive = true
}

// Pause deactivates the timer
func (c *Countdown) Pause() {
	c.IsActive = false
}

// End resets and pauses the timer
func (c *Countdown) End() {
	c.Pause()
	c.Reset()
}

// GetPercentRemaining returns the percentage of time remaining
func (c *Countdown) GetPercentRemaining() int {
	return int(100 * (float64(c.RemainingTime) / float64(c.OriginalDuration)))
}

// FormatTime returns the remaining time formatted as HH:MM:SS
func (c *Countdown) FormatTime() string {
	hours := int(c.RemainingTime.Hours())
	minutes := int(c.RemainingTime.Minutes()) % 60
	seconds := int(c.RemainingTime.Seconds()) % 60
	
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
