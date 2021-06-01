# S029

The S029 analyzer reports cases of schemas which enables only `Computed`
and configures `ExactlyOneOf`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    ExactlyOneOf: []string{"example"},
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S029` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S029
&schema.Schema{
    Computed:     true,
    ExactlyOneOf: []string{"example"},
}
```
