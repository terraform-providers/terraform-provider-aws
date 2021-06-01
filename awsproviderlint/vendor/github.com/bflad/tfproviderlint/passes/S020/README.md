# S020

The S020 analyzer reports cases of schemas which enables only `Computed`
and enables `ForceNew`, which is invalid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    ForceNew: true,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S020` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S020
&schema.Schema{
    Computed: true,
    ForceNew: true,
}
```
