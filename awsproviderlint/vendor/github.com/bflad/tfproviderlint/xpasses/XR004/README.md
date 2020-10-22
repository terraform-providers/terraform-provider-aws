# XR004

The XR004 analyzer reports [`Set()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set) calls that receive a complex value type, but do not perform error checking. This error checking is to prevent issues where the code is not able to properly set the Terraform state for drift detection. Addition details are available in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/best-practices/detecting-drift.html#error-checking-aggregate-types).

## Flagged Code

```go
d.Set("example", []interface{}{})

d.Set("example", map[string]interface{}{})

d.Set("example", schema.NewSet(/* ... */))
```

## Passing Code

```go
if err := d.Set("example", []interface{}{}); err != nil {
    return fmt.Errorf("error setting example: %s", err)
}

if err := d.Set("example", map[string]interface{}{}); err != nil {
    return fmt.Errorf("error setting example: %s", err)
}

if err := d.Set("example", schema.NewSet(/* ... */)); err != nil {
    return fmt.Errorf("error setting example: %s", err)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR004` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR004
d.Set("example", []interface{}{})
```
