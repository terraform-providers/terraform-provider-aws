package tagresource

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Separator used in resource identifiers
	ResourceIdSeparator = `,`
)

// GetResourceId parses a given resource identifier for tag identifier and tag key.
func GetResourceId(resourceId string) (string, string, error) {
	parts := strings.SplitN(resourceId, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid resource identifier (%s), expected ID,KEY", resourceId)
	}

	return parts[0], parts[1], nil
}

// SetResourceId creates a resource identifier given a tag identifier and a tag key.
func SetResourceId(identifier string, key string) string {
	return identifier + ResourceIdSeparator + key
}

// toSnakeCase converts a string to snake case.
//
// For example, AWS Go SDK field names are in PascalCase,
// while Terraform schema attribute names are in snake_case.
func toSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}
