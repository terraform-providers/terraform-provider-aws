# S012

The S012 analyzer reports cases of schemas which `Type`
is not configured, which will fail provider schema validation.

## Flagged Code

```go
_ = schema.Schema{
    Computed: true,
}

_ = schema.Schema{
    Optional: true,
}

_ = schema.Schema{
    Required: true,
}
```

## Passing Code

```go
_ = schema.Schema{
    Computed: true,
    Type:     schema.TypeString,
}

_ = schema.Schema{
    Optional: true,
    Type:     schema.TypeString,
}

_ = schema.Schema{
    Required: true,
    Type:     schema.TypeString,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S012` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S012
_ = schema.Schema{
    Computed: true,
}
```
