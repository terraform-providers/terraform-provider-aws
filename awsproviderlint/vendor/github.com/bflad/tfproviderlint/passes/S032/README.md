# S032

The S032 analyzer reports cases of schemas which enables only `Computed`
and configures `MinItems`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    MinItems: 1,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S032` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S032
&schema.Schema{
    Computed: true,
    MinItems: 1,
}
```
