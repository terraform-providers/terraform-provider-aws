# S010

The S010 analyzer reports cases of schemas which enables only `Computed`
and configures `ValidateFunc`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    ValidateFunc: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S010` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S010
&schema.Schema{
    Computed:     true,
    ValidateFunc: /* ... */,
}
```
