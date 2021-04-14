# R004

The R004 analyzer reports incorrect types for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) call value.
The `Set()` function only supports a subset of basic types, slices and maps of that
subset of basic types, and the [`schema.Set`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#Set) type.

## Flagged Code

```go
var t time.Time

d.Set("example", t)
```

## Passing Code

```go
var t time.Time

d.Set("example", t.Format(time.RFC3339))
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R004` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
var t time.Time

//lintignore:R004
d.Set("example", t)
```
