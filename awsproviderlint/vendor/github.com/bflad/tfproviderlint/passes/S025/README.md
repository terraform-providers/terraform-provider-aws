# S025

The S025 analyzer reports cases of schemas which enables only `Computed`
and configures `AtLeastOneOf`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    AtLeastOneOf: []string{"example"},
    Computed:     true,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S025` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S025
&schema.Schema{
    AtLeastOneOf: []string{"example"},
    Computed:     true,
}
```
