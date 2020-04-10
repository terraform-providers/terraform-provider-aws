# S003

The S003 analyzer reports cases of schemas which enables both `Required`
and `Computed`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required: true,
    Computed: true,
}
```

## Passing Code

```go
&schema.Schema{
    Required: true,
}

# OR

&schema.Schema{
    Optional: true,
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S003` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S003
&schema.Schema{
    Required: true,
    Computed: true,
}
```
