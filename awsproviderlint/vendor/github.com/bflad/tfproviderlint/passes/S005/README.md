# S005

The S005 analyzer reports cases of schemas which enables `Computed`
and configures `Default`, which will fail provider schema validation.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    Default:  /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}

# OR

&schema.Schema{
    Default: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S005` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S005
&schema.Schema{
    Computed: true,
    Default:  /* ... */,
}
```
