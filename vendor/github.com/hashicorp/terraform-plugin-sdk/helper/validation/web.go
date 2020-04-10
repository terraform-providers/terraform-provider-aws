package validation

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// IsURLWithHTTPS is a SchemaValidateFunc which tests if the provided value is of type string and a valid HTTPS URL
func IsURLWithHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return IsURLWithScheme([]string{"https"})(i, k)
}

// IsURLWithHTTPorHTTPS is a SchemaValidateFunc which tests if the provided value is of type string and a valid HTTP or HTTPS URL
func IsURLWithHTTPorHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return IsURLWithScheme([]string{"http", "https"})(i, k)
}

// IsURLWithScheme is a SchemaValidateFunc which tests if the provided value is of type string and a valid URL with the provided schemas
func IsURLWithScheme(validSchemes []string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
			return
		}

		if v == "" {
			errors = append(errors, fmt.Errorf("expected %q url to not be empty, got %v", k, i))
			return
		}

		u, err := url.Parse(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("expected %q to be a valid url, got %v: %+v", k, v, err))
			return
		}

		if u.Host == "" {
			errors = append(errors, fmt.Errorf("expected %q to have a host, got %v", k, v))
			return
		}

		for _, s := range validSchemes {
			if u.Scheme == s {
				return //last check so just return
			}
		}

		errors = append(errors, fmt.Errorf("expected %q to have a url with schema of: %q, got %v", k, strings.Join(validSchemes, ","), v))
		return
	}
}
