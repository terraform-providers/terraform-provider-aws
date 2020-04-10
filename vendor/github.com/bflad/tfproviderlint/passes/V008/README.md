# V008

The V008 analyzer reports usage of the deprecated [ValidateRFC3339TimeString](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#ValidateRFC3339TimeString) validation function that should be replaced with [IsRFC3339Time](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#IsRFC3339Time).

## Flagged Code

```go
ValidateFunc: validation.ValidateRFC3339TimeString,
```

## Passing Code

```go
ValidateFunc: validation.IsRFC3339Time,
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V008` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V008
ValidateFunc: validation.ValidateRFC3339TimeString,
```
