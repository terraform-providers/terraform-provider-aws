# V011

The V011 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with [validation.StringLenBetween()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringLenBetween).

## Flagged Code

```go
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if len(value) < 1 || len(value) > 255 {
    errors = append(errors, fmt.Errorf("%q must be between 1 and 255 characters: %q", k, value))
  }

  return
}
```

## Passing Code

```go
// directly in Schema
ValidateFunc: validation.StringLenBetween(1, 255),

// or saving as a variable
var validateExampleThing = validation.StringLenBetween(1, 255)

// or replacing the function
func validateExampleThing() schema.SchemaValidateFunc {
  return validation.StringLenBetween(1, 255)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V011` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V011
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if len(value) < 1 || len(value) > 255 {
    errors = append(errors, fmt.Errorf("%q must be between 1 and 255 characters: %q", k, value))
  }

  return
}
```
