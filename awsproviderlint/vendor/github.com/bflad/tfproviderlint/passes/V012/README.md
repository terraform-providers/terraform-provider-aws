# V012

The V012 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with [validation.IntAtLeast()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#IntAtLeast), [validation.IntAtMost()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#IntAtMost), or [validation.IntBetween()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#IntBetween).

## Flagged Code

```go
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(int)

  if value < 1 || value > 255 {
    errors = append(errors, fmt.Errorf("%q must be between 1 and 255: %d", k, value))
  }

  return
}
```

## Passing Code

```go
// directly in Schema
ValidateFunc: validation.IntBetween(1, 255),

// or saving as a variable
var validateExampleThing = validation.IntBetween(1, 255)

// or replacing the function
func validateExampleThing() schema.SchemaValidateFunc {
  return validation.IntBetween(1, 255)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V012` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V012
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(int)

  if value < 1 || value > 255 {
    errors = append(errors, fmt.Errorf("%q must be between 1 and 255: %d", k, value))
  }

  return
}
```
