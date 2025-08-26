package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	configs "github.com/DeveloperPaul123/agenda/internal/configs"
	models "github.com/DeveloperPaul123/agenda/internal/models"
	providers "github.com/DeveloperPaul123/agenda/internal/providers"
	spinner "github.com/briandowns/spinner"
)

// Version and commit will be set during build time using -ldflags
var (
	version = "dev"
	commit  = "none"
)

// EventFormatter handles formatting events for output to the console.
type EventFormatter struct {
	timeFormat    string
	eventTemplate *template.Template
}

// NewEventFormatter creates a new EventFormatter with the given time format and event template string.
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

// FormatEvent formats a CalendarEvent using the configured template and time format.
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

// initConfig initializes the default configuration file.
func initConfig(cmd *cobra.Command, args []string) {
	config := configs.DefaultConfig()
	if err := configs.WriteConfig(config); err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	fmt.Printf("Created default configuration file at: %s\n", configs.DefaultConfigPath())
	fmt.Printf("Please set your API key in the %s environment variable.\n", config.Providers[config.Provider].EnvAPIKey)
}

// runAgenda is the main function that runs the agenda command.
func runAgenda(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config")
	provider, _ := cmd.Flags().GetString("provider")
	timeFormat, _ := cmd.Flags().GetString("time-format")
	eventTemplate, _ := cmd.Flags().GetString("event-template")
	verbose, _ := cmd.Flags().GetBool("verbose")
	dateStr, _ := cmd.Flags().GetString("date")

	if configPath == "" {
		configPath = configs.DefaultConfigPath()
	}

	config, err := configs.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if provider != "" {
		config.Provider = provider
	}
	if timeFormat != "" {
		config.TimeFormat = timeFormat
	}
	if eventTemplate != "" {
		config.EventTemplate = eventTemplate
	}

	useDate := time.Now()
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatalf("Invalid date format: %v. Use YYYY-MM-DD.", err)
		}
		useDate = parsedDate
	}

	if verbose {
		log.Printf("Using provider: %s", config.Provider)
		log.Printf("Time format: %s", config.TimeFormat)
		log.Printf("Event template: %s", config.EventTemplate)
	}

	factory := providers.NewProviderFactory(config)
	calProvider, err := factory.CreateProvider(config.Provider)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Start()

	events, err := calProvider.GetTodaysEvents(useDate)
	s.Stop()
	if err != nil {
		log.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		fmt.Println("No events found for today.")
		return
	}

	uniqueEvents := make(map[string]models.CalendarEvent)
	for _, event := range events {
		key := fmt.Sprintf("%s-%s", event.Title, event.StartTime.Format(time.RFC3339))
		if _, exists := uniqueEvents[key]; !exists {
			uniqueEvents[key] = event
		}
	}

	if len(uniqueEvents) == 0 {
		fmt.Println("No events today")
		return
	}

	// Convert uniqueEvents map back to a list
	sortedEvents := make([]models.CalendarEvent, 0, len(uniqueEvents))
	for _, event := range uniqueEvents {
		sortedEvents = append(sortedEvents, event)
	}

	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].StartTime.Local().Before(sortedEvents[j].StartTime.Local())
	})

	formatter, err := NewEventFormatter(config.TimeFormat, config.EventTemplate)
	if err != nil {
		log.Fatalf("Failed to create formatter: %v", err)
	}

	for _, event := range sortedEvents {
		formatted, err := formatter.FormatEvent(event)
		if err != nil {
			log.Printf("Warning: failed to format event %s: %v", event.Title, err)
			continue
		}
		fmt.Println(formatted)
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "agenda",
		Short:   "Agenda CLI",
		Run:     runAgenda,
		Version: fmt.Sprintf("%s-%s", version, commit),
	}

	// We don't need completions
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Disable the help subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Define flags
	rootCmd.Flags().String("config", "", "Path to configuration file (default: ~/.config/agenda/config.yaml)")
	rootCmd.Flags().String("provider", "", "Override the provider from config")
	rootCmd.Flags().String("time-format", "", "Override the time format from config")
	rootCmd.Flags().String("event-template", "", "Override the event template from config")
	rootCmd.Flags().Bool("verbose", false, "Enable verbose logging")
	rootCmd.Flags().String("date", "", "Date to get events for (format: YYYY-MM-DD, default is today)")

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration",
		Run:   initConfig,
	}
	rootCmd.AddCommand(initCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
