# S027

The S027 analyzer reports cases of schemas which enables only `Computed`
and configures `Default`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    Default:  "example",
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S027` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S027
&schema.Schema{
    Computed: true,
    Default:  "example",
}
```
