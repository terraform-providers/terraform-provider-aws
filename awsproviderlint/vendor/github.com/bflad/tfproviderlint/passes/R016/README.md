# R016

The R016 analyzer reports [`(*schema.ResourceData).SetId()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) usage with unstable `resource.PrefixedUniqueId()` value. Schema attributes should be stable across Terraform runs.

## Flagged Code

```go
d.SetId(resource.PrefixedUniqueId("example"))
```

## Passing Code

```go
d.SetId("stablestring")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R016` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R016
d.SetId(resource.PrefixedUniqueId("example"))
```
