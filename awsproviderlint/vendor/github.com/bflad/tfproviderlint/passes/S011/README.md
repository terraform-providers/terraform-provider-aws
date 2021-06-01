# S011

The S011 analyzer reports cases of schemas which enables only `Computed`
and configures `DiffSuppressFunc`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed:         true,
    DiffSuppressFunc: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S011` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S011
&schema.Schema{
    Computed:         true,
    DiffSuppressFunc: /* ... */,
}
```
