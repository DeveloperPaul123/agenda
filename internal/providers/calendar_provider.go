package providers

import (
	models "github.com/DeveloperPaul123/agenda/internal/models"
)

// CalendarProvider interface for different calendar services
type CalendarProvider interface {
	GetTodaysEvents() ([]models.CalendarEvent, error)
	GetName() string
}
