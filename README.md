# Remi - Reminder Application

A terminal-based reminder application.

## Project Structure

The application has been refactored to follow the standard Go project structure:

```
remi/
├── cmd/
│   └── remi/           # Main application entry point
│       └── main.go
├── internal/           # Application-specific packages
│   ├── app/            # Main application logic
│   ├── config/         # Configuration handling
│   ├── freeze/         # Freeze mode functionality
│   ├── model/          # Business logic and application state
│   └── ui/             # Terminal user interface
├── pkg/                # Reusable packages
│   └── timer/          # Countdown timer functionality
└── remi.yaml           # Configuration file
```

## Features

- Multiple configurable reminders
- Terminal-based user interface
- Start, pause, and reset timers
- Dialog-based freeze mode to enforce breaks (macOS only)

## Configuration

The application is configured via the `remi.yaml` file. The configuration file is loaded in the following order:

1. First, it tries to load from the home directory (`~/remi.yaml`)
2. If not found in the home directory, it falls back to the current directory

Example configuration:

```yaml
config:
  # Application-wide settings
  # Currently no global settings are used

events:
  - name: "Drink water"
    interval: "30m"
    notify: true        # Legacy option, kept for backward compatibility
    autoStart: true     # Optional, defaults to true if not specified
    freeze: "1m"        # Optional, shows dialog for 1 minute when timer completes
  - name: "Stand up and walk around"
    interval: "1h"
    notify: true        # Legacy option, kept for backward compatibility
    autoStart: false    # Optional, set to false to start in paused state
    freeze: "5m"        # Optional, shows dialog for 5 minutes when timer completes
```

### Freeze Functionality

The `freeze` option allows you to enforce breaks by displaying a blocking dialog when a timer completes:

- When a timer with `freeze` set completes, a macOS dialog appears for the specified duration
- During the freeze period, the timer shows "⏸ Freezing (MM:SS)" with a countdown of the remaining time
- You cannot control (start/pause/reset) a timer while it's in freeze mode
- After the freeze period ends, the timer automatically starts the next round

This is useful for enforcing breaks and preventing you from immediately dismissing reminders.

## Commands

- `s <index>` - Start/resume a timer
- `p <index>` - Pause a timer
- `e <index>` - End/reset a timer

## Running the Application

```bash
go run cmd/remi/main.go
```

Or build and run:

```bash
go build -o remi cmd/remi/main.go
./remi
```

## Design Principles

The refactored code follows these design principles:

1. **Separation of Concerns**: Each package has a specific responsibility
2. **Dependency Injection**: Components are created and passed where needed
3. **Interface-Based Design**: Using interfaces for better testability
4. **Error Handling**: Improved error handling throughout the application
5. **Scalability**: Structure makes it easier to add new features
