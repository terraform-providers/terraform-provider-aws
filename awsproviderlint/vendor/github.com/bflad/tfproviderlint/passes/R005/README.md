# R005

The R005 analyzer reports when multiple [`HasChange()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.HasChange) calls in a conditional can be combined into a single [`HasChanges()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.HasChanges) call.

## Flagged Code

```go
if d.HasChange("attr1") || d.HasChange("attr2") {
  // handle attr1 and attr2 changes
}
```

## Passing Code

```go
if d.HasChanges("attr1", "attr2") {
  // handle attr1 and attr2 changes
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R005` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R005
if d.HasChange("attr1") || d.HasChange("attr2") {
  // handle attr1 and attr2 changes
}
```
