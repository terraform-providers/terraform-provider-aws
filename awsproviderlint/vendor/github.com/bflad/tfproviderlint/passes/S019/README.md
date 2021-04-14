# S019

The S019 analyzer reports cases of schemas including `Computed: false`,
`Optional: false`, or `Required: false` which are extraneous.

## Flagged Code

```go
&schema.Schema{
    Computed: false,
    Optional: true,
}

&schema.Schema{
    Optional: false,
    Required: true,
}

&schema.Schema{
    Computed: true,
    Required: false,
}
```

## Passing Code

```go
&schema.Schema{
    Optional: true,
}

&schema.Schema{
    Required: true,
}

&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S019` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S019
&schema.Schema{
    Computed: false,
    Optional: true,
}
```
