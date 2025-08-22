package providers

import (
	"time"

	models "github.com/DeveloperPaul123/agenda/internal/models"
)

// CalendarProvider interface for different calendar services
type CalendarProvider interface {
	GetTodaysEvents(date time.Time) ([]models.CalendarEvent, error)
	GetName() string
}
