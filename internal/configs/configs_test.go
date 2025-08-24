package configs

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	defaultConfig := DefaultConfig()
	if defaultConfig.Provider == "" {
		t.Error("Default provider should not be empty")
	}
	if defaultConfig.TimeFormat == "" {
		t.Error("Default time format should not be empty")
	}
	if defaultConfig.EventTemplate == "" {
		t.Error("Default event template should not be empty")
	}
	if len(defaultConfig.Providers) == 0 {
		t.Error("Default providers should not be empty")
	}
}
