# Remi - Reminder Application

A terminal-based reminder application with system notifications support.

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
│   ├── model/          # Business logic and application state
│   ├── notification/   # System notification handling
│   └── ui/             # Terminal user interface
├── pkg/                # Reusable packages
│   └── timer/          # Countdown timer functionality
└── remi.yaml           # Configuration file
```

## Features

- Multiple configurable reminders
- System notifications (macOS supported)
- Terminal-based user interface
- Start, pause, and reset timers
- Fullscreen freeze mode to enforce breaks (macOS only)

## Configuration

The application is configured via the `remi.yaml` file. The configuration file is loaded in the following order:

1. First, it tries to load from the home directory (`~/remi.yaml`)
2. If not found in the home directory, it falls back to the current directory

Example configuration:

```yaml
config:
  use_system_notification: true

events:
  - name: "Drink water"
    interval: "30m"
    notify: true
    autoStart: true  # Optional, defaults to true if not specified
    freeze: "1m"     # Optional, locks screen for 1 minute when timer completes
  - name: "Stand up and walk around"
    interval: "1h"
    notify: true
    autoStart: false  # Optional, set to false to start in paused state
    freeze: "5m"      # Optional, locks screen for 5 minutes when timer completes
```

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
