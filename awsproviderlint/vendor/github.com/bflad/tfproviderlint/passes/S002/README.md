# S002

The S002 analyzer reports cases of schemas which enables both `Required`
and `Optional`, which will fail provider schema validation.

## Flagged Code

```go
_ = schema.Schema{
    Required: true,
    Optional: true,
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
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S002
_ = schema.Schema{
    Required: true,
    Optional: true,
}
```
