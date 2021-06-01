# V002

The V002 analyzer reports usage of the deprecated [CIDRNetwork](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#CIDRNetwork) validation function that should be replaced with [IsCIDRNetwork](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#IsCIDRNetwork).

## Flagged Code

```go
ValidateFunc: validation.CIDRNetwork(0, 32),
```

## Passing Code

```go
ValidateFunc: validation.IsCIDRNetwork(0, 32),
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V002
ValidateFunc: validation.CIDRNetwork(0, 32),
```
