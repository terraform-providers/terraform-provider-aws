# R007

The R007 analyzer reports usage of the deprecated [(helper/schema.ResourceData).Partial()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/schema?tab=doc#ResourceData.Partial) function that does not need replacement.

## Flagged Code

```go
d.Partial(true),
```

## Passing Code

```go
// Not present :)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R007` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R007
d.Partial(true),
```
