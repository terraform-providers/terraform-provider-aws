package validation

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

// StringIsNotEmpty is a ValidateFunc that ensures a string is not empty
func StringIsNotEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string, got %v", k, i)}
	}

	return nil, nil
}

// StringIsNotWhiteSpace is a ValidateFunc that ensures a string is not empty or consisting entirely of whitespace characters
func StringIsNotWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string or whitespace", k)}
	}

	return nil, nil
}

// StringIsEmpty is a ValidateFunc that ensures a string has no characters
func StringIsEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string: got %v", k, v)}
	}

	return nil, nil
}

// StringIsWhiteSpace is a ValidateFunc that ensures a string is composed of entirely whitespace
func StringIsWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string or whitespace: got %v", k, v)}
	}

	return nil, nil
}

// StringLenBetween returns a SchemaValidateFunc which tests if the provided value
// is of type string and has length between min and max (inclusive)
func StringLenBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		if len(v) < min || len(v) > max {
			errors = append(errors, fmt.Errorf("expected length of %s to be in the range (%d - %d), got %s", k, min, max, v))
		}

		return warnings, errors
	}
}

// StringMatch returns a SchemaValidateFunc which tests if the provided value
// matches a given regexp. Optionally an error message can be provided to
// return something friendlier than "must match some globby regexp".
func StringMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
		}

		if ok := r.MatchString(v); !ok {
			if message != "" {
				return nil, []error{fmt.Errorf("invalid value for %s (%s)", k, message)}

			}
			return nil, []error{fmt.Errorf("expected value of %s to match regular expression %q, got %v", k, r, i)}
		}
		return nil, nil
	}
}

// StringDoesNotMatch returns a SchemaValidateFunc which tests if the provided value
// does not match a given regexp. Optionally an error message can be provided to
// return something friendlier than "must not match some globby regexp".
func StringDoesNotMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
		}

		if ok := r.MatchString(v); ok {
			if message != "" {
				return nil, []error{fmt.Errorf("invalid value for %s (%s)", k, message)}

			}
			return nil, []error{fmt.Errorf("expected value of %s to not match regular expression %q, got %v", k, r, i)}
		}
		return nil, nil
	}
}

// StringInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type string and matches the value of an element in the valid slice
// will test with in lower case if ignoreCase is true
func StringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		for _, str := range valid {
			if v == str || (ignoreCase && strings.ToLower(v) == strings.ToLower(str)) {
				return warnings, errors
			}
		}

		errors = append(errors, fmt.Errorf("expected %s to be one of %v, got %s", k, valid, v))
		return warnings, errors
	}
}

// StringNotInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type string and does not match the value of any element in the invalid slice
// will test with in lower case if ignoreCase is true
func StringNotInSlice(invalid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		for _, str := range invalid {
			if v == str || (ignoreCase && strings.ToLower(v) == strings.ToLower(str)) {
				errors = append(errors, fmt.Errorf("expected %s to not be any of %v, got %s", k, invalid, v))
				return warnings, errors
			}
		}

		return warnings, errors
	}
}

// StringDoesNotContainAny returns a SchemaValidateFunc which validates that the
// provided value does not contain any of the specified Unicode code points in chars.
func StringDoesNotContainAny(chars string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		if strings.ContainsAny(v, chars) {
			errors = append(errors, fmt.Errorf("expected value of %s to not contain any of %q, got %v", k, chars, i))
			return warnings, errors
		}

		return warnings, errors
	}
}

// StringIsBase64 is a ValidateFunc that ensures a string can be parsed as Base64
func StringIsBase64(i interface{}, k string) (warnings []string, errors []error) {
	// Empty string is not allowed
	if warnings, errors = StringIsNotEmpty(i, k); len(errors) > 0 {
		return
	}

	// NoEmptyStrings checks it is a string
	v, _ := i.(string)

	if _, err := base64.StdEncoding.DecodeString(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a base64 string, got %v", k, v))
	}

	return warnings, errors
}

// ValidateJsonString is a SchemaValidateFunc which tests to make sure the
// supplied string is valid JSON.
//
// Deprecated: use StringIsJSON instead
func ValidateJsonString(i interface{}, k string) (warnings []string, errors []error) {
	return StringIsJSON(i, k)
}

// StringIsJSON is a SchemaValidateFunc which tests to make sure the supplied string is valid JSON.
func StringIsJSON(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if _, err := structure.NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}

	return warnings, errors
}

// ValidateRegexp returns a SchemaValidateFunc which tests to make sure the
// supplied string is a valid regular expression.
//
// Deprecated: use StringIsValidRegExp instead
func ValidateRegexp(i interface{}, k string) (warnings []string, errors []error) {
	return StringIsValidRegExp(i, k)
}

// StringIsValidRegExp returns a SchemaValidateFunc which tests to make sure the supplied string is a valid regular expression.
func StringIsValidRegExp(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if _, err := regexp.Compile(v); err != nil {
		errors = append(errors, fmt.Errorf("%q: %s", k, err))
	}

	return warnings, errors
}
