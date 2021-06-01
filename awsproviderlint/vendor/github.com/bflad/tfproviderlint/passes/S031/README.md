# S031

The S031 analyzer reports cases of schemas which enables only `Computed`
and configures `MaxItems`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    MaxItems: 1,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S031` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S031
&schema.Schema{
    Computed: true,
    MaxItems: 1,
}
```
