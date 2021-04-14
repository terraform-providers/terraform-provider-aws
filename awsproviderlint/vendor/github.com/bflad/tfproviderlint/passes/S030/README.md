# S030

The S030 analyzer reports cases of schemas which enables only `Computed`
and configures `InputDefault`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    InputDefault: "example",
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S030` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S030
&schema.Schema{
    Computed:     true,
    InputDefault: "example",
}
```
