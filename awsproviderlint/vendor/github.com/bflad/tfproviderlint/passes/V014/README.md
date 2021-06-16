# V014

The V014 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with [validation.IntInSlice()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#IntInSlice) or [validation.IntNotInSlice()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#IntNotInSlice).

## Flagged Code

```go
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(int)

  if value != 1 && value != 2 {
    errors = append(errors, fmt.Errorf("%q must be 1 or 2: %d", k, value))
  }

  return
}
```

## Passing Code

```go
// directly in Schema
ValidateFunc: validation.IntInSlice([]int{1, 2}, false),

// or saving as a variable
var validateExampleThing = validation.IntInSlice([]int{1, 2}, false)

// or replacing the function
func validateExampleThing() schema.SchemaValidateFunc {
  return validation.IntInSlice([]int{1, 2}, false)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V014` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V014
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(int)

  if value != 1 && value != 2 {
    errors = append(errors, fmt.Errorf("%q must be 1 or 2: %d", k, value))
  }

  return
}
```
