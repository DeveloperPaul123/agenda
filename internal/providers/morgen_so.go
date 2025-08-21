package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/DeveloperPaul123/agenda/internal/configs"
	models "github.com/DeveloperPaul123/agenda/internal/models"
	duration "github.com/channelmeter/iso8601duration"
)

// MorgenProvider implements CalendarProvider for Morgen.so
type MorgenProvider struct {
	config configs.ProviderConfig
	apiKey string
}

// morgenCalenderRights represents the rights a user has on a calendar in Morgen
// API response. It is used to determine if the user can read items in the calendar.
type morgenCalenderRights struct {
	CanRead bool `json:"mayReadItems"`
}

// morgenCalendar represents a calendar in the Morgen API response.
type morgenCalendar struct {
	Name           string               `json:"name"`
	AccountId      string               `json:"accountId"`
	CalenderRights morgenCalenderRights `json:"myRights"`
	Id             string               `json:"id"`
	Color          string               `json:"color"`
}

// morgenCalendarsResponseData represents the response structure from Morgen API
// for the list of calendars. It contains a slice of morgenCalendar objects.
type morgenCalendarsResponseData struct {
	Calendars []morgenCalendar `json:"calendars"`
	// TODO: Accounts?
}

// morgenCalendarsResponse represents the response structure from Morgen API
// for the list of calendars. It contains a data field with morgenCalendarsResponseData.
type morgenCalendarsResponse struct {
	Data morgenCalendarsResponseData `json:"data"`
}

// morgenEvent represents the response structure from Morgen API
type morgenEvent struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StartTime   string `json:"start"`
	Duration    string `json:"duration"`
	TimeZone    string `json:"timeZone"`
	EndTime     string `json:"end"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// morgenEventsResponseData represents the response structure from Morgen API
// for the list of events. It contains a slice of morgenEvent objects.
type morgenEventsResponseData struct {
	Events []morgenEvent `json:"events"`
}

// morgenEventsResponse represents the response structure from Morgen API
// for the list of events. It contains a data field with morgenEventsResponseData.
type morgenEventsResponse struct {
	Data morgenEventsResponseData `json:"data"`
}

// morgenProviderName is the name of the Morgen provider.
// It is used to identify the provider in the application.
const morgenProviderName = "morgen"

// contains checks if a string is present in a slice of strings.
func contains(list []string, target string) bool {
	return slices.Contains(list, target)
}

// ProviderName returns the name of the Morgen provider.
func ProviderName() string {
	return morgenProviderName
}

// NewMorgenProvider creates a new instance of MorgenProvider with the given configuration.
func NewMorgenProvider(config configs.ProviderConfig) *MorgenProvider {
	return &MorgenProvider{
		config: config,
		apiKey: os.Getenv(config.EnvAPIKey),
	}
}

// GetName returns the name of the provider.
func (m *MorgenProvider) GetName() string {
	return ProviderName()
}

// getApiKey retrieves the API key from the environment variable specified in the provider configuration.
// If the API key is not set, it returns an error.
func (m *MorgenProvider) getApiKey() (string, error) {
	if m.apiKey == "" {
		return "", fmt.Errorf("API key not found in environment variable %s", m.config.EnvAPIKey)
	}

	return m.apiKey, nil
}

// getCalendars retrieves the list of calendars from the Morgen API along with account info but we currently only use the calender data response.
// Returns a list of morgenCalendar objects or an error if the request fails.
func (m *MorgenProvider) getCalendars() ([]morgenCalendar, error) {
	apiKey, err := m.getApiKey()
	if err != nil {
		return nil, err
	}

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
			value = strings.Replace(value, "{API_KEY}", apiKey, 1)
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

	// Parse response
	var responseData morgenCalendarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return responseData.Data.Calendars, nil
}

// GetTodaysEvents retrieves today's events from the Morgen API.
func (m *MorgenProvider) GetTodaysEvents() ([]models.CalendarEvent, error) {
	apiKey, err := m.getApiKey()
	if err != nil {
		return nil, err
	}

	calendars, err := m.getCalendars()
	if err != nil {
		return nil, err
	}

	accountCalendarMap := make(map[string][]string)
	for i := range calendars {
		cal := calendars[i]
		// Only include calendars that the user has read access to and are not in the ignore list
		if cal.CalenderRights.CanRead && !contains(m.config.CalendarsToIgnore, cal.Name) {
			accountCalendarMap[cal.AccountId] = append(accountCalendarMap[cal.AccountId], cal.Id)
		}
	}

	// Get today's date range
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Build URL with date range
	url := fmt.Sprintf("%s/events/list",
		m.config.BaseURL)

	// Loop through key values and create requests for each one and run in parallel
	var morgenEvents []morgenEvent
	for accountId, calendarIds := range accountCalendarMap {
		// Create request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		query := req.URL.Query()
		query.Set("start", startOfDay.Format(time.RFC3339))
		query.Set("end", endOfDay.Format(time.RFC3339))

		query.Set("accountId", accountId)
		query.Set("calendarIds", strings.Join(calendarIds, ","))
		req.URL.RawQuery = query.Encode()

		// Add headers
		for key, value := range m.config.Headers {
			if key == "Authorization" {
				value = strings.Replace(value, "{API_KEY}", apiKey, 1)
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

		// Parse response
		var response morgenEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		morgenEvents = append(morgenEvents, response.Data.Events...)
	}

	// Convert to standard format
	var events []models.CalendarEvent
	for _, me := range morgenEvents {
		loc, err := time.LoadLocation(me.TimeZone)
		if err != nil {
			log.Printf("Warning: failed to load timezone %s: %v", me.TimeZone, err)
			continue
		}

		// Response times do not have the timezone, that is a separate field
		startTime, err := time.ParseInLocation("2006-01-02T15:04:05", me.StartTime, loc)
		if err != nil {
			log.Printf("Warning: failed to parse start time %s: %v", me.StartTime, err)
			continue
		}

		// Parse the duration to get the end time
		dur, err := duration.FromString(me.Duration)
		if err != nil {
			log.Printf("Warning: failed to parse duration %s: %v", me.Duration, err)
		}
		endTime := startTime.Add(dur.ToDuration())

		if err != nil {
			log.Printf("Warning: failed to parse end time %s: %v", me.EndTime, err)
			continue
		}

		events = append(events, models.CalendarEvent{
			ID:    me.ID,
			Title: me.Title,
			// Convert start and end times to the correct timezone
			StartTime:   startTime.In(time.Local),
			EndTime:     endTime.In(time.Local),
			Description: me.Description,
			Location:    me.Location,
		})
	}

	return events, nil
}
