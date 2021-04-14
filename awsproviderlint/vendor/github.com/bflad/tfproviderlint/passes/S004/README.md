# S004

The S004 analyzer reports cases of schemas which enables `Required`
and configures `Default`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Required: true,
    Default:  /* ... */,
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
    Default:  /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S004` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S004
&schema.Schema{
    Required: true,
    Default:  /* ... */,
}
```
