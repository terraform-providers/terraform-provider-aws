# S014

The S014 analyzer reports cases of schemas which the `Elem` has `Computed`,
`Optional`, or `Required` configured, which will fail provider schema validation.

## Flagged Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Elem:     &schema.Schema{
            Required: true,
            Type:     schema.TypeString,
        },
        Required: true,
        Type:     schema.TypeList,
    },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Elem:     &schema.Schema{
            Type: schema.TypeString,
        },
        Required: true,
        Type:     schema.TypeList,
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S014` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S014
map[string]*schema.Schema{
    "attribute_name": {
        Elem:     &schema.Schema{
            Required: true,
            Type:     schema.TypeString,
        },
        Required: true,
        Type:     schema.TypeList,
    },
}
```
