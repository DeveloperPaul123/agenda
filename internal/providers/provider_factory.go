package providers

import (
	"fmt"

	. "github.com/DeveloperPaul123/agenda/internal/configs"
)

// ProviderFactory creates calendar providers
type ProviderFactory struct {
	config Config
}

func NewProviderFactory(config Config) *ProviderFactory {
	return &ProviderFactory{config: config}
}

func (f *ProviderFactory) CreateProvider(name string) (CalendarProvider, error) {
	providerConfig, exists := f.config.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found in configuration", name)
	}

	switch name {
	case "morgen":
		return NewMorgenProvider(providerConfig), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
