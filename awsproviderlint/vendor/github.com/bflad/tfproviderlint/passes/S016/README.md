# S016

The S016 analyzer reports cases of schemas including `Set` without `TypeSet`,
which will fail schema validation.

## Flagged Code

```go
&schema.Schema{
    Set:  /* ... */,
    Type: schema.TypeList,
}
```

## Passing Code

```go
&schema.Schema{
    Set:  /* ... */,
    Type: schema.TypeSet,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S016` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S016
&schema.Schema{
    Set:  /* ... */,
    Type: schema.TypeList,
}
```
