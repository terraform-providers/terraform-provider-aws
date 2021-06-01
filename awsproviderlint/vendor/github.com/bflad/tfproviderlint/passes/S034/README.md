# S034

The S034 analyzer reports cases of schemas which enable `PromoteSingle`, which is not valid after Terraform 0.12. Existing implementations of `PromoteSingle` prior to Terraform 0.12 can be ignored currently.

## Flagged Code

```go
&schema.Schema{
    PromoteSingle: true,
}
```

## Passing Code

```go
&schema.Schema{
    // No PromoteSingle: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S034` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S034
&schema.Schema{
    PromoteSingle: true,
}
```
