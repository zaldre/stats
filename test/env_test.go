package test

import (
	"os"
	"testing"

	"github.com/snowpea/stats/internal/config"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		envValue   string
		expected   string
		shouldSet  bool
	}{
		{
			name:       "Environment variable exists",
			key:        "TEST_VAR",
			defaultVal: "default",
			envValue:   "env_value",
			expected:   "env_value",
			shouldSet:  true,
		},
		{
			name:       "Environment variable does not exist",
			key:        "NONEXISTENT_VAR",
			defaultVal: "default",
			envValue:   "",
			expected:   "default",
			shouldSet:  false,
		},
		{
			name:       "Environment variable is empty",
			key:        "EMPTY_VAR",
			defaultVal: "default",
			envValue:   "",
			expected:   "default",
			shouldSet:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing environment variable
			os.Unsetenv(tt.key)

			if tt.shouldSet {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := config.GetEnv(tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("GetEnv(%s, %s) = %s, expected %s", tt.key, tt.defaultVal, result, tt.expected)
			}
		})
	}
}
