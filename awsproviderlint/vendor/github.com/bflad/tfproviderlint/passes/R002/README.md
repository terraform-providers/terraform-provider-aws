# R002

The R002 analyzer reports likely extraneous uses of
star (`*`) dereferences for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) call. The `Set()` function automatically
handles pointers and `*` dereferences without `nil` checks can panic.

## Flagged Code

```go
var stringPtr *string

d.Set("example", *stringPtr)
```

## Passing Code

```go
var stringPtr *string

d.Set("example", stringPtr)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
var stringPtr *string

//lintignore:R002
d.Set("example", *stringPtr)
```
