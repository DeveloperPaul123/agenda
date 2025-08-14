package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DeveloperPaul123/agenda/internal/configs"
	models "github.com/DeveloperPaul123/agenda/internal/models"
)

// MorgenProvider implements CalendarProvider for Morgen.so
type MorgenProvider struct {
	config configs.ProviderConfig
	apiKey string
}

type MorgenCalendar struct {
	Name      string `json:"name"`
	AccountId string `json:"accountId"`
}

// MorgenEvent represents the response structure from Morgen API
type MorgenEvent struct {
	ID          string `json:"id"`
	Title       string `json:"summary"`
	StartTime   string `json:"start"`
	EndTime     string `json:"end"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

func NewMorgenProvider(config configs.ProviderConfig) *MorgenProvider {
	return &MorgenProvider{
		config: config,
		apiKey: os.Getenv(config.EnvAPIKey),
	}
}

func (m *MorgenProvider) GetName() string {
	return "morgen"
}

func (m *MorgenProvider) GetTodaysEvents() ([]models.CalendarEvent, error) {
	if m.apiKey == "" {
		return nil, fmt.Errorf("API key not found in environment variable %s", m.config.EnvAPIKey)
	}

	// Get today's date range
	// now := time.Now()
	// startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// endOfDay := startOfDay.Add(24 * time.Hour)

	// Build URL with date range
	url := fmt.Sprintf("%s/calendars/list",
		m.config.BaseURL)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range m.config.Headers {
		if key == "Authorization" {
			value = strings.Replace(value, "{API_KEY}", m.apiKey, 1)
		}
		req.Header.Set(key, value)
	}

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// TODO(PT): Parse the available calendars from the response.
	// Group these by account and then send requests for all the calendars that are part of each unique account ID

	// Parse response
	var morgenEvents []MorgenEvent
	if err := json.NewDecoder(resp.Body).Decode(&morgenEvents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to standard format
	var events []models.CalendarEvent
	for _, me := range morgenEvents {
		startTime, err := time.Parse(time.RFC3339, me.StartTime)
		if err != nil {
			log.Printf("Warning: failed to parse start time %s: %v", me.StartTime, err)
			continue
		}

		endTime, err := time.Parse(time.RFC3339, me.EndTime)
		if err != nil {
			log.Printf("Warning: failed to parse end time %s: %v", me.EndTime, err)
			continue
		}

		events = append(events, models.CalendarEvent{
			ID:          me.ID,
			Title:       me.Title,
			StartTime:   startTime,
			EndTime:     endTime,
			Description: me.Description,
			Location:    me.Location,
		})
	}

	return events, nil
}