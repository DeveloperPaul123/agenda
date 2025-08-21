package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v3"
)

const CURRENT_CONFIG_VERSION uint64 = 1
const CONFIG_FILE_NAME string = "agenda.conf"
const CONFIG_FOLDER string = "agenda"

// Config represents the application configuration
type Config struct {
	Provider      string                    `yaml:"provider"`
	TimeFormat    string                    `yaml:"time_format"`
	EventTemplate string                    `yaml:"event_template"`
	Providers     map[string]ProviderConfig `yaml:"providers"`
	Version       uint64                    `yaml:"config_version"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	BaseURL           string            `yaml:"base_url"`
	Headers           map[string]string `yaml:"headers"`
	EnvAPIKey         string            `yaml:"env_api_key"`
	CalendarsToIgnore []string          `yaml:"calendars_to_ignore"`
}

// Returns the default configuration for the application.
func DefaultConfig() Config {
	// Default configuration for now
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
				EnvAPIKey:         "MORGEN_API_KEY",
				CalendarsToIgnore: []string{"ignore_this_calendar"},
			},
		},
		Version: CURRENT_CONFIG_VERSION,
	}

	return config
}

// DefaultConfigPath returns the default path for the configuration file.
func DefaultConfigPath() string {
	return getSystemConfigPath()
}

// WriteConfig writes the provided configuration to the system's config directory.
// It creates the directory if it does not exist.
func WriteConfig(config Config) error {
	configPath := getSystemConfigPath()
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

// ReadConfig reads the configuration from the specified path.
// If the file does not exist, it creates a default configuration and writes it to the path.
// If the version of the configuration does not match the current version, it tries to merge the configuration with the default one and writes it back.
// Returns the configuration and any error encountered.
func ReadConfig(path string) (Config, error) {
	var config Config
	configFile := path

	//Does the file exist?
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		WriteConfig(config)
		return config, err
	} else {
		if err := loadConfig(configFile, &config); err != nil {
			return config, err
		}

		if config.Version != CURRENT_CONFIG_VERSION {
			// TODO: Actually migrate config versions
			config = DefaultConfig()
			WriteConfig(config)
		}
	}

	return config, nil
}

// loadConfig loads the configuration from the specified file path.
func loadConfig(configPath string, config *Config) error {
	// Try to load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return nil
}

func getSystemConfigPath() string {
	// Factor out the constant
	return getConfigPath(configdir.LocalCache(CONFIG_FOLDER))
}

func getConfigPath(configDir string) string {
	err := configdir.MakePath(configDir)
	if err != nil {
		panic(err)
	}

	configFile := filepath.Join(configDir, CONFIG_FILE_NAME)
	return configFile
}
