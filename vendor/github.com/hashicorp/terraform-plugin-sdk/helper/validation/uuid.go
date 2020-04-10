package validation

import (
	"fmt"

	"github.com/hashicorp/go-uuid"
)

// IsUUID is a ValidateFunc that ensures a string can be parsed as UUID
func IsUUID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := uuid.ParseUUID(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid UUID, got %v", k, v))
	}

	return warnings, errors
}
