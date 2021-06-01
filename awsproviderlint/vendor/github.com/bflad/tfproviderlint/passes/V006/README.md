# V006

The V006 analyzer reports usage of the deprecated [ValidateListUniqueStrings](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#ValidateListUniqueStrings) validation function that should be replaced with [ListOfUniqueStrings](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/validation#ListOfUniqueStrings).

## Flagged Code

```go
ValidateFunc: validation.ValidateListUniqueStrings,
```

## Passing Code

```go
ValidateFunc: validation.ListOfUniqueStrings,
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:V006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:V006
ValidateFunc: validation.ValidateListUniqueStrings,
```
