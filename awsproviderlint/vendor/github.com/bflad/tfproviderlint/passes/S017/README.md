# S017

The S017 analyzer reports cases of schemas including `MaxItems` or
`MinItems` without `TypeList`, `TypeMap`, or `TypeSet`, which will
fail schema validation.

## Flagged Code

```go
&schema.Schema{
    MaxItems: 1,
    Type:     schema.TypeString,
}

&schema.Schema{
    MinItems: 1,
    Type:     schema.TypeString,
}
```

## Passing Code

```go
&schema.Schema{
    Type: schema.TypeString,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S017` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S017
&schema.Schema{
    MaxItems: 1,
    Type:     schema.TypeString,
}
```
