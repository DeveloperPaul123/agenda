package configs

// Config represents the application configuration
type Config struct {
	Provider      string                    `yaml:"provider"`
	TimeFormat    string                    `yaml:"time_format"`
	EventTemplate string                    `yaml:"event_template"`
	Providers     map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	BaseURL   string            `yaml:"base_url"`
	Headers   map[string]string `yaml:"headers"`
	EnvAPIKey string            `yaml:"env_api_key"`
}

func DefaultConfig() Config {
	// Default configuration
	config := Config{
		Provider:      "morgen",
		TimeFormat:    "15:04",
		EventTemplate: "- {{.StartTimeFormatted}}-{{.EndTimeFormatted}}: {{.Title}}",
		Providers: map[string]ProviderConfig{
			"morgen": {
				BaseURL: "https://api.morgen.so/v3",
				Headers: map[string]string{
					"Authorization": "ApiKey {API_KEY}",
					"Content-Type":  "application/json",
				},
				EnvAPIKey: "MORGEN_API_KEY",
			},
		},
	}

	return config
}