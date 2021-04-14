# S007

The S007 analyzer reports cases of schemas which enables `Required`
and configures `ConflictsWith`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required:      true,
    ConflictsWith: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Required: true,
}

# OR

&schema.Schema{
    Optional:      true,
    ConflictsWith: /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S007` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S007
&schema.Schema{
    Required:      true,
    ConflictsWith: /* ... */,
}
```
