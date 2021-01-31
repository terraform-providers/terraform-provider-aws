# S026

The S026 analyzer reports cases of schemas which enables only `Computed`
and configures `ConflictsWith`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed:      true,
    ConflictsWith: []string{"example"},
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S026` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S026
&schema.Schema{
    Computed:      true,
    ConflictsWith: []string{"example"},
}
```
