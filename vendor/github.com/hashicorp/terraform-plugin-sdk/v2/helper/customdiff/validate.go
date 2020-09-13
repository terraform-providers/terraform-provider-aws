package customdiff

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ValueChangeValidationFunc is a function type that validates the difference
// (or lack thereof) between two values, returning an error if the change
// is invalid.
type ValueChangeValidationFunc func(ctx context.Context, old, new, meta interface{}) error

// ValueValidationFunc is a function type that validates a particular value,
// returning an error if the value is invalid.
type ValueValidationFunc func(ctx context.Context, value, meta interface{}) error

// ValidateChange returns a CustomizeDiffFunc that applies the given validation
// function to the change for the given key, returning any error produced.
func ValidateChange(key string, f ValueChangeValidationFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		old, new := d.GetChange(key)
		return f(ctx, old, new, meta)
	}
}

// ValidateValue returns a CustomizeDiffFunc that applies the given validation
// function to value of the given key, returning any error produced.
//
// This should generally not be used since it is functionally equivalent to
// a validation function applied directly to the schema attribute in question,
// but is provided for situations where composing multiple CustomizeDiffFuncs
// together makes intent clearer than spreading that validation across the
// schema.
func ValidateValue(key string, f ValueValidationFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		val := d.Get(key)
		return f(ctx, val, meta)
	}
}
