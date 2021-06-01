# XS001

The XS001 analyzer reports cases of schemas where `Description` is not configured, which is generally useful for providers that wish to automatically generate documentation based on the schema information.

## Flagged Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Optional: true,
        Type:     schema.TypeString,
    },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Description: "does something useful",
        Optional:    true,
        Type:        schema.TypeString,
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XS001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XS001
map[string]*schema.Schema{
    "attribute_name": {
        Optional: true,
        Type:     schema.TypeString,
    },
}
```
