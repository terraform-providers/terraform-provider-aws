# V013

The V013 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with [validation.StringInSlice()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringInSlice) or [validation.StringNotInSlice()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringNotInSlice).

## Flagged Code

```go
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if value != "value1" && value != "value2" {
    errors = append(errors, fmt.Errorf("%q must be value1 or value2: %s", k, value))
  }

  return
}
```

## Passing Code

```go
// directly in Schema
ValidateFunc: validation.StringInSlice([]string{"value1", "value2"}, false),

// or saving as a variable
var validateExampleThing = validation.StringInSlice([]string{"value1", "value2"}, false)

// or replacing the function
func validateExampleThing() schema.SchemaValidateFunc {
  return validation.StringInSlice([]string{"value1", "value2"}, false)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V013` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V013
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if value != "value1" && value != "value2" {
    errors = append(errors, fmt.Errorf("%q must be value1 or value2: %s", k, value))
  }

  return
}
```
