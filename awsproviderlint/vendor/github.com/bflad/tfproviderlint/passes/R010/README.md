# R010

The R010 analyzer reports when [(helper/schema.ResourceData).GetChange()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/schema?tab=doc#ResourceData.GetChange) assignments are not using the first return value (assigned to `_`), which should be replaced with [(helper/schema.ResourceData).Get()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/schema?tab=doc#ResourceData.Get) instead.

## Flagged Code

```go
_, n := d.GetChange("example")
```

## Passing Code

```go
n := d.Get("example")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R010` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R010
_, n := d.GetChange("example")
```
