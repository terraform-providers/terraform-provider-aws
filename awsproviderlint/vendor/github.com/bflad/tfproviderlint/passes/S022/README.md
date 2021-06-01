# S022

The S022 analyzer reports cases of schema that declare `Elem` of `*schema.Resource`
with `TypeMap`, which has undefined behavior. Only `TypeList` and `TypeSet` can be
used for configuration block attributes.

## Flagged Code

```go
&schema.Schema{
  Type: schema.TypeMap,
  Elem: &schema.Resource{},
}
```

## Passing Code

```go
&schema.Schema{
  Type: schema.TypeList,
  Elem: &schema.Resource{},
}

// or

&schema.Schema{
  Type: schema.TypeSet,
  Elem: &schema.Resource{},
}

// or

&schema.Schema{
  Type: schema.TypeMap,
  Elem: &schema.Schema{},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S022` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S022
&schema.Schema{
  Type: schema.TypeMap,
  Elem: &schema.Resource{},
}
```
