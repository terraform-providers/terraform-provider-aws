# V009

The V009 analyzer reports when the second argument for a [`validation.StringMatch()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringMatch) call is an empty string. It is preferred to provide a friendly validation message, rather than allowing the function to return the raw regular expression as the message, since not all practitioners may be familiar with regular expression syntax.

## Flagged Code

```go
validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9.-]+$`), "")
```

## Passing Code

```go
validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9.-]+$`), "must contain only alphanumeric characters, periods, or hyphens")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V009` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9.-]+$`), "") //lintignore:V009
```
