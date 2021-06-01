# R008

The R008 analyzer reports usage of the deprecated [(helper/schema.ResourceData).SetPartial()](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/schema?tab=doc#ResourceData.SetPartial) function that does not need replacement.

## Flagged Code

```go
d.SetPartial("example"),
```

## Passing Code

```go
// Not present :)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R008` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R008
d.SetPartial("example"),
```
