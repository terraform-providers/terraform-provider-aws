# S013

The S013 analyzer reports cases of schemas which one of `Computed`,
`Optional`, or `Required` is not configured, which will fail provider
schema validation.

## Flagged Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Type: schema.TypeString,
    },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "attribute_name": {
        Computed: true,
        Type:     schema.TypeString,
    },
}

# OR

map[string]*schema.Schema{
    "attribute_name": {
        Optional: true,
        Type:     schema.TypeString,
    },
}

# OR

map[string]*schema.Schema{
    "attribute_name": {
        Required: true,
        Type:     schema.TypeString,
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S013` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S013
map[string]*schema.Schema{
    "attribute_name": {
        Type: schema.TypeString,
    },
}
```
