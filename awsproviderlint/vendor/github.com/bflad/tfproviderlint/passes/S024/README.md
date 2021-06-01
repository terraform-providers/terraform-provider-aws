# S024

The S024 analyzer reports extraneous usage of `ForceNew` in data source schema attributes.

## Flagged Code

```go
&schema.Resource{
  Read: /* ... */,
  Schema: map[string]*schema.Schema{
    "example": {
      ForceNew: true,
      Required: true,
      Type:     schema.TypeString,
    },
  },
}
```

## Passing Code

```go
&schema.Resource{
  Read: /* ... */,
  Schema: map[string]*schema.Schema{
    "example": {
      Required: true,
      Type:     schema.TypeString,
    },
  },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S024` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S024
&schema.Resource{
  Read: /* ... */,
  Schema: map[string]*schema.Schema{
    "example": {
      ForceNew: true,
      Required: true,
      Type:     schema.TypeString,
    },
  },
}
```
