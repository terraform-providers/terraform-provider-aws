package envvar

import (
	"fmt"
	"os"
)

// GetWithDefault gets an environment variable value if non-empty or returns the default.
func GetWithDefault(variable string, defaultValue string) string {
	value := os.Getenv(variable)

	if value == "" {
		return defaultValue
	}

	return value
}

// RequireOneOf verifies that at least one environment variable is non-empty or returns an error.
//
// If at lease one environment variable is non-empty, returns the first name and value.
func RequireOneOf(names []string, usageMessage string) (string, string, error) {
	for _, variable := range names {
		value := os.Getenv(variable)

		if value != "" {
			return variable, value, nil
		}
	}

	return "", "", fmt.Errorf("at least one environment variable of %v must be set. Usage: %s", names, usageMessage)
}

// Require verifies that an environment variable is non-empty or returns an error.
func Require(name string, usageMessage string) (string, error) {
	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value, nil
}
