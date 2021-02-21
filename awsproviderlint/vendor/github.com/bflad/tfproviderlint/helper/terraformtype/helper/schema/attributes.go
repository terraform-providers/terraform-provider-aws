package schema

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Pattern for schema attribute names
	AttributeNameRegexpPattern = `^[a-z0-9_]+$`

	// Pattern for schema references to attributes, such as ConflictsWith values
	AttributeReferenceRegexpPattern = `^[a-z0-9_]+(\.[a-z0-9_]+)*$`
)

var (
	AttributeNameRegexp      = regexp.MustCompile(AttributeNameRegexpPattern)
	AttributeReferenceRegexp = regexp.MustCompile(AttributeReferenceRegexpPattern)
)

// ParseAttributeReference validates and returns the split representation of schema attribute reference.
// Attribute references are used in Schema fields such as AtLeastOneOf, ConflictsWith, and ExactlyOneOf.
func ParseAttributeReference(reference string) ([]string, error) {
	if !AttributeReferenceRegexp.MatchString(reference) {
		return nil, fmt.Errorf("%q must contain only valid attribute names, separated by periods", reference)
	}

	attributeReferenceParts := strings.Split(reference, ".")

	if len(attributeReferenceParts) == 1 {
		return attributeReferenceParts, nil
	}

	configurationBlockReferenceErr := fmt.Errorf("%q configuration block attribute references are only valid for TypeList and MaxItems: 1 attributes and nested attributes must be separated by .0.", reference)

	if math.Mod(float64(len(attributeReferenceParts)), 2) == 0 {
		return attributeReferenceParts, configurationBlockReferenceErr
	}

	// All even parts of an attribute reference must be 0
	for idx, attributeReferencePart := range attributeReferenceParts {
		if math.Mod(float64(idx), 2) == 0 {
			continue
		}

		attributeReferencePartInt, err := strconv.Atoi(attributeReferencePart)

		if err != nil {
			return attributeReferenceParts, configurationBlockReferenceErr
		}

		if attributeReferencePartInt != 0 {
			return attributeReferenceParts, configurationBlockReferenceErr
		}
	}

	return attributeReferenceParts, nil
}
