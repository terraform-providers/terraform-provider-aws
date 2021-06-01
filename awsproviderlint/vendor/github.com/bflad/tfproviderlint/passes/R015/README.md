# R015

The R015 analyzer reports [`(*schema.ResourceData).SetId()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) usage with unstable `resource.UniqueId()` value. Schema attributes should be stable across Terraform runs.

## Flagged Code

```go
d.SetId(resource.UniqueId())
```

## Passing Code

```go
d.SetId("stablestring")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R015` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R015
d.SetId(resource.UniqueId())
```
