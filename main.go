package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	configs "github.com/DeveloperPaul123/agenda/internal/configs"
	models "github.com/DeveloperPaul123/agenda/internal/models"
	providers "github.com/DeveloperPaul123/agenda/internal/providers"
)

// EventFormatter handles formatting events for output
type EventFormatter struct {
	timeFormat    string
	eventTemplate *template.Template
}

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

// loadConfig loads configuration from file
func loadConfig(configPath string) (configs.Config, error) {
	config := configs.DefaultConfig()

	// Try to load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}

// createDefaultConfigFile creates a default config file
func createDefaultConfigFile(configPath string) error {
	config := configs.DefaultConfig()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func main() {
	var (
		configPath    = flag.String("config", "", "Path to configuration file (default: ~/.config/agenda/config.yaml)")
		initConfig    = flag.Bool("init", false, "Create a default configuration file")
		provider      = flag.String("provider", "", "Override the provider from config")
		timeFormat    = flag.String("time-format", "", "Override the time format from config")
		eventTemplate = flag.String("event-template", "", "Override the event template from config")
		verbose       = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Set default config path
	if *configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		*configPath = filepath.Join(homeDir, ".config", "agenda", "config.yaml")
	}

	// Initialize config if requested
	if *initConfig {
		if err := createDefaultConfigFile(*configPath); err != nil {
			log.Fatalf("Failed to create config file: %v", err)
		}
		fmt.Printf("Created default configuration file at: %s\n", *configPath)
		fmt.Println("Please set your API key in the environment variable specified in the config.")
		return
	}

	// Load configuration
	config, err := loadConfig(*configPath)
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

	// Get today's events
	events, err := calProvider.GetTodaysEvents()
	if err != nil {
		log.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		fmt.Println("No events found for today.")
		return
	}

	// Create formatter
	formatter, err := NewEventFormatter(config.TimeFormat, config.EventTemplate)
	if err != nil {
		log.Fatalf("Failed to create formatter: %v", err)
	}

	// Format and output events
	fmt.Printf("# Today's Meetings (%s)\n\n", time.Now().Format("January 2, 2006"))
	for _, event := range events {
		formatted, err := formatter.FormatEvent(event)
		if err != nil {
			log.Printf("Warning: failed to format event %s: %v", event.Title, err)
			continue
		}
		fmt.Println(formatted)
	}
}
