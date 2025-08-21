package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	configs "github.com/DeveloperPaul123/agenda/internal/configs"
	models "github.com/DeveloperPaul123/agenda/internal/models"
	providers "github.com/DeveloperPaul123/agenda/internal/providers"
	spinner "github.com/briandowns/spinner"
)

// EventFormatter handles formatting events for output to the console.
type EventFormatter struct {
	timeFormat    string
	eventTemplate *template.Template
}

// NewEventFormatter creates a new EventFormatter with the specified time format
// and event template string. It returns an error if the template parsing fails.
func NewEventFormatter(timeFormat, eventTemplateStr string) (*EventFormatter, error) {
	tmpl, err := template.New("event").Parse(eventTemplateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event template: %w", err)
	}

	return &EventFormatter{
		timeFormat:    timeFormat,
		eventTemplate: tmpl,
	}, nil
}

// FormatEvent formats a calendar event into a string using the configured template
// and time format. It returns the formatted string or an error if formatting fails.
func (f *EventFormatter) FormatEvent(event models.CalendarEvent) (string, error) {
	data := struct {
		models.CalendarEvent
		StartTimeFormatted string
		EndTimeFormatted   string
		Duration           string
	}{
		CalendarEvent:      event,
		StartTimeFormatted: event.StartTime.Format(f.timeFormat),
		EndTimeFormatted:   event.EndTime.Format(f.timeFormat),
		Duration:           event.EndTime.Sub(event.StartTime).String(),
	}

	var result strings.Builder
	if err := f.eventTemplate.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// initConfig initializes the default configuration file if it does not exist.
func initConfig() {
	config := configs.DefaultConfig()

	if err := configs.WriteConfig(config); err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	fmt.Printf("Created default configuration file at: %s\n", configs.DefaultConfigPath())
	fmt.Printf("Please set your API key in the %s environment variable.\n", config.Providers[config.Provider].EnvAPIKey)
}

func main() {
	// Manually handle subcommands since we only support 1
	if len(os.Args) >= 2 {
		subcommand := strings.TrimSpace(strings.ToLower(os.Args[1]))
		// is this a flag or a subcommand?
		if subcommand[0] != '-' {
			switch subcommand {
			case "init":
				initConfig()
				os.Exit(0)
			default:
				fmt.Printf("Unknown subcommand: %s\n", subcommand)
				os.Exit(1)
			}
		}
	}

	// Flags that we support
	var (
		configPath    = flag.String("config", "", "Path to configuration file (default: ~/.config/agenda/config.yaml)")
		provider      = flag.String("provider", "", "Override the provider from config")
		timeFormat    = flag.String("time-format", "", "Override the time format from config")
		eventTemplate = flag.String("event-template", "", "Override the event template from config")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Set default config path
	if *configPath == "" {
		*configPath = configs.DefaultConfigPath()
	}

	// Load configuration
	config, err := configs.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Apply command line overrides
	if *provider != "" {
		config.Provider = *provider
	}
	if *timeFormat != "" {
		config.TimeFormat = *timeFormat
	}
	if *eventTemplate != "" {
		config.EventTemplate = *eventTemplate
	}

	if *verbose {
		log.Printf("Using provider: %s", config.Provider)
		log.Printf("Time format: %s", config.TimeFormat)
		log.Printf("Event template: %s", config.EventTemplate)
	}

	// Create provider
	factory := providers.NewProviderFactory(config)
	calProvider, err := factory.CreateProvider(config.Provider)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Start()
	// Get today's events
	events, err := calProvider.GetTodaysEvents()
	s.Stop()
	if err != nil {
		log.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		fmt.Println("No events found for today.")
		return
	}

	// Sort events by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime.In(time.Local).Before(events[j].StartTime.In(time.Local))
	})

	// Deduplicate events by title and time
	uniqueEvents := make(map[string]models.CalendarEvent)
	for _, event := range events {
		key := fmt.Sprintf("%s-%s", event.Title, event.StartTime.Format(time.RFC3339))
		if _, exists := uniqueEvents[key]; !exists {
			uniqueEvents[key] = event
		}
	}

	// Create formatter
	formatter, err := NewEventFormatter(config.TimeFormat, config.EventTemplate)
	if err != nil {
		log.Fatalf("Failed to create formatter: %v", err)
	}

	// Format and output events
	for _, event := range uniqueEvents {
		formatted, err := formatter.FormatEvent(event)
		if err != nil {
			log.Printf("Warning: failed to format event %s: %v", event.Title, err)
			continue
		}
		fmt.Println(formatted)
	}
}
