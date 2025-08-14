# Agenda CLI

A command-line tool that pulls calendar events from various calendar providers and formats them for easy copy-pasting into markdown documents.

## Features

- Support for multiple calendar providers (currently Morgen.so, extensible for others)
- Configurable time formatting
- Customizable event templates using Go templates
- Configuration via YAML file
- Environment variable support for API keys
- Clean markdown output

## Installation

1. Clone or download the source code
2. Run `go mod tidy` to install dependencies
3. Build the application: `go build -o agenda`

## Quick Start

1. Initialize a default configuration:
   ```bash
   ./agenda --init
   ```

2. Set your API key in the environment:
   ```bash
   export MORGEN_API_KEY="your-morgen-api-key"
   ```

3. Run the agenda:
   ```bash
   ./agenda
   ```

## Configuration

The configuration file is located at `~/.config/agenda/config.yaml` by default.

### Example Configuration

```yaml
provider: morgen
time_format: "15:04"
event_template: "- {{.StartTimeFormatted}}-{{.EndTimeFormatted}}: {{.Title}}"

providers:
  morgen:
    base_url: "https://api.morgen.so/v1"
    headers:
      Authorization: "Bearer {API_KEY}"
      Content-Type: "application/json"
    env_api_key: "MORGEN_API_KEY"
```

### Configuration Options

- `provider`: Which calendar provider to use (e.g., "morgen")
- `time_format`: Go time format string for displaying times (e.g., "15:04", "3:04 PM")
- `event_template`: Go template string for formatting events

### Template Variables

The following variables are available in the event template:

- `{{.Title}}` - Event title
- `{{.StartTimeFormatted}}` - Start time formatted with your time_format
- `{{.EndTimeFormatted}}` - End time formatted with your time_format
- `{{.Duration}}` - Duration of the event
- `{{.Description}}` - Event description
- `{{.Location}}` - Event location
- `{{.ID}}` - Event ID

### Example Templates

```yaml
# Simple format
event_template: "- {{.StartTimeFormatted}}: {{.Title}}"

# Detailed format with location
event_template: "- {{.StartTimeFormatted}}-{{.EndTimeFormatted}}: **{{.Title}}**{{if .Location}} ({{.Location}}){{end}}"

# With duration
event_template: "- {{.StartTimeFormatted}} ({{.Duration}}): {{.Title}}"
```

## Command Line Options

- `--config PATH` - Specify a custom configuration file path
- `--init` - Create a default configuration file
- `--provider NAME` - Override the provider from config
- `--time-format FORMAT` - Override the time format from config
- `--event-template TEMPLATE` - Override the event template from config
- `--verbose` - Enable verbose logging

## Environment Variables

- `MORGEN_API_KEY` - Your Morgen.so API key

## Adding New Providers

To add support for a new calendar provider:

1. Implement the `CalendarProvider` interface:
   ```go
   type CalendarProvider interface {
       GetTodaysEvents() ([]CalendarEvent, error)
       GetName() string
   }
   ```

2. Add the provider to the `CreateProvider` function in the `ProviderFactory`

3. Add the provider configuration to the default config

## Output Example

```markdown
# Today's Meetings (December 15, 2024)

- 09:00-10:00: Team Standup
- 10:30-11:30: Project Review
- 14:00-15:00: Client Call
- 16:00-16:30: 1:1 with Manager
```

## API Key Setup

### Morgen.so

1. Log into your Morgen account
2. Go to Settings > Integrations > API
3. Generate a new API key
4. Set it as the `MORGEN_API_KEY` environment variable

## Troubleshooting

- Ensure your API key is correctly set in the environment
- Check that the provider configuration matches your API requirements
- Use `--verbose` flag for detailed logging
- Verify your internet connection and API endpoint accessibility

## License
