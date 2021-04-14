# V003

The V003 analyzer reports usage of the deprecated [IPRange](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#IPRange) validation function that should be replaced with [IsIPv4Range](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#IsIPv4Range).

## Flagged Code

```go
ValidateFunc: validation.IPRange(),
```

## Passing Code

```go
ValidateFunc: validation.IsIPv4Range,
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V003` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V003
ValidateFunc: validation.IPRange(),
```
