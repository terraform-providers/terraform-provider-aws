# V010

The V010 analyzer reports when the second argument for a [`validation.StringDoesNotMatch()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation#StringDoesNotMatch) call is an empty string. It is preferred to provide a friendly validation message, rather than allowing the function to return the raw regular expression as the message, since not all practitioners may be familiar with regular expression syntax.

## Flagged Code

```go
validation.StringDoesNotMatch(regexp.MustCompile(`^[!@#$%^&*()]+$`), "")
```

## Passing Code

```go
validation.StringDoesNotMatch(regexp.MustCompile(`^[!@#$%^&*()]+$`), "must not contain exclamation, at, octothorp, US dollar, percentage, carat, ampersand, star, or parenthesis symbols")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V010` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
validation.StringDoesNotMatch(regexp.MustCompile(`^[!@#$%^&*()]+$`), "") //lintignore:V010
```
