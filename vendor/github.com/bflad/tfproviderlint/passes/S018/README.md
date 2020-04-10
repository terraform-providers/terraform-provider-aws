# S018

The S018 analyzer reports cases of `Schema` including `MaxItems: 1` and `Type: schema.TypeSet` that should be simplified to `Type: schema.TypeList`.

## Flagged Code

```go
&schema.Schema{
    MaxItems: 1,
    Type:     schema.TypeSet,
}
```

## Passing Code

```go
&schema.Schema{
    MaxItems: 1,
    Type:     schema.TypeList,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S018` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S018
&schema.Schema{
    MaxItems: 1,
    Type:     schema.TypeSet,
}
```
