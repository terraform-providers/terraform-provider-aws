# S001

The S001 analyzer reports cases of `TypeList` or `TypeSet` schemas missing `Elem`,
which will fail schema validation.

## Flagged Code

```go
&schema.Schema{
    Type: schema.TypeList,
}

&schema.Schema{
    Type: schema.TypeSet,
}
```

## Passing Code

```go
&schema.Schema{
    Type: schema.TypeList,
    Elem: &schema.Schema{Type: schema.TypeString},
}

&schema.Schema{
    Type: schema.TypeSet,
    Elem: &schema.Schema{Type: schema.TypeString},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S001
&schema.Schema{
    Type: schema.TypeList,
}
```
