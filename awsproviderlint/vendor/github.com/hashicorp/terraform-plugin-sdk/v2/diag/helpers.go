package diag

import "fmt"

// FromErr will convert an error into a Diagnostics. This returns Diagnostics
// as the most common use case in Go will be handling a single error
// returned from a function.
//
//   if err != nil {
//     return diag.FromErr(err)
//   }
func FromErr(err error) Diagnostics {
	if err == nil {
		return nil
	}
	return Diagnostics{
		Diagnostic{
			Severity: Error,
			Summary:  err.Error(),
		},
	}
}

// Errorf creates a Diagnostics with a single Error level Diagnostic entry.
// The summary is populated by performing a fmt.Sprintf with the supplied
// values. This returns a single error in a Diagnostics as errors typically
// do not occur in multiples as warnings may.
//
//   if unexpectedCondition {
//     return diag.Errorf("unexpected: %s", someValue)
//   }
func Errorf(format string, a ...interface{}) Diagnostics {
	return Diagnostics{
		Diagnostic{
			Severity: Error,
			Summary:  fmt.Sprintf(format, a...),
		},
	}
}
