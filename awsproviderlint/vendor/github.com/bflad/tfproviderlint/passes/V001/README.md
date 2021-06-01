# V001

The V001 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with [validation.StringMatch()](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#StringMatch) or [validation.StringDoesNotMatch()](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#StringDoesNotMatch).

## Flagged Code

```go
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if !regexp.MustCompile(`^s-([0-9a-f]{17})$`).MatchString(value) {
    errors = append(errors, fmt.Errorf("%q must begin with s- and contain 17 lowercase alphanumeric characters: %q", k, value))
  }

  return
}
```

## Passing Code

```go
// directly in Schema
ValidateFunc: validation.StringMatch(regexp.MustCompile(`^s-([0-9a-f]{17})$`), "must begin with s- and contain 17 lowercase alphanumeric characters"),

// or saving as a variable
var validateExampleThing = validation.StringMatch(regexp.MustCompile(`^s-([0-9a-f]{17})$`), "must begin with s- and contain 17 lowercase alphanumeric characters")

// or replacing the function
func validateExampleThing() schema.SchemaValidateFunc {
  return validation.StringMatch(regexp.MustCompile(`^s-([0-9a-f]{17})$`), "must begin with s- and contain 17 lowercase alphanumeric characters")
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V001
func validateExampleThing(v interface{}, k string) (ws []string, errors []error) {
  value := v.(string)

  if !regexp.MustCompile(`^s-([0-9a-f]{17})$`).MatchString(value) {
    errors = append(errors, fmt.Errorf("%q must begin with s- and contain 17 lowercase alphanumeric characters: %q", k, value))
  }

  return
}
```
