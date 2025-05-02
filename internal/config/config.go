// Package config handles the application configuration loading and validation
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration loaded from remi.yaml
type Config struct {
	Events    []Event          `yaml:"events"`
	AppConfig AppConfiguration `yaml:"config"`
}

// AppConfiguration stores general application settings
type AppConfiguration struct {
	UseSystemNotification bool `yaml:"use_system_notification"`
}

// Event represents a timed event with notification settings
type Event struct {
	Name      string `yaml:"name"`
	Interval  string `yaml:"interval"`
	Notify    bool   `yaml:"notify"`
	AutoStart *bool  `yaml:"autoStart,omitempty"`
	Freeze    string `yaml:"freeze,omitempty"`
}

// EventConfig represents a parsed event with validated duration
type EventConfig struct {
	Name             string
	OriginalDuration time.Duration
	UseNotification  bool
	IsActive         bool
	FreezeDuration   time.Duration
}

// Load reads and parses the configuration file
func Load(filename string) (Config, error) {
	var config Config

	// Try to load from home directory first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeConfig := filepath.Join(homeDir, filename)
		if data, err := os.ReadFile(homeConfig); err == nil {
			if err := yaml.Unmarshal(data, &config); err == nil {
				// Successfully loaded from home directory
				return validateConfig(config)
			}
		}
	}

	// Fall back to current directory if not found in home directory
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("cannot read file %s (tried home directory and current directory): %w", filename, err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("cannot parse YAML: %w", err)
	}

	return validateConfig(config)
}

// validateConfig validates the loaded configuration
func validateConfig(config Config) (Config, error) {
	// Validate configuration
	if len(config.Events) == 0 {
		return config, fmt.Errorf("no events found in configuration file")
	}

	return config, nil
}

// ParseEvents converts raw events from config to validated EventConfig objects
func ParseEvents(events []Event) ([]EventConfig, error) {
	result := make([]EventConfig, 0, len(events))
	
	for _, event := range events {
		duration, err := time.ParseDuration(event.Interval)
		if err != nil {
			return nil, fmt.Errorf("error parsing duration '%s' for event '%s': %w",
				event.Interval, event.Name, err)
		}

		// Handle default value for autoStart if not specified
		autoStart := true
		if event.AutoStart != nil && *event.AutoStart == false {
			autoStart = false
		}

		// Parse freeze duration if specified
		var freezeDuration time.Duration
		if event.Freeze != "" {
			freezeDuration, err = time.ParseDuration(event.Freeze)
			if err != nil {
				return nil, fmt.Errorf("error parsing freeze duration '%s' for event '%s': %w",
					event.Freeze, event.Name, err)
			}
		}

		result = append(result, EventConfig{
			Name:             event.Name,
			OriginalDuration: duration,
			UseNotification:  event.Notify,
			IsActive:         autoStart,
			FreezeDuration:   freezeDuration,
		})
	}

	return result, nil
}
