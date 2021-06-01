# S015

The S015 analyzer reports cases of schemas which the attribute name
includes characters outside lowercase alphanumerics and underscores,
which will fail provider schema validation.

## Flagged Code

```go
map[string]*schema.Schema{
    "INVALID": {
        Required: true,
        Type:     schema.TypeString,
    },
}

map[string]*schema.Schema{
    "invalid!": {
        Required: true,
        Type:     schema.TypeString,
    },
}

map[string]*schema.Schema{
    "invalid-name": {
        Required: true,
        Type:     schema.TypeString,
    },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "valid": {
        Required: true,
        Type:     schema.TypeString,
    },
}

map[string]*schema.Schema{
    "valid_name": {
        Required: true,
        Type:     schema.TypeString,
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S015` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S015
map[string]*schema.Schema{
    "INVALID": {
        Required: true,
        Type:     schema.TypeString,
    },
}
```
