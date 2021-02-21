# R017

The R017 analyzer reports [`(*schema.ResourceData).SetId()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) usage with unstable `time.Now()` value. Schema attributes should be stable across Terraform runs.

## Flagged Code

```go
d.SetId(time.Now().Format(time.RFC3339))
```

## Passing Code

```go
d.SetId("stablestring")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R017` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R017
d.SetId(time.Now().Format(time.RFC3339))
```
