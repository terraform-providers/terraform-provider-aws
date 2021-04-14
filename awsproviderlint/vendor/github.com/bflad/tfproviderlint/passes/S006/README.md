# S006

The S006 analyzer reports cases of `TypeMap` schemas missing `Elem`,
which currently passes Terraform schema validation, but breaks downstream tools
and may be required in the future.

## Flagged Code

```go
&schema.Schema{
    Type: schema.TypeMap,
}
```

## Passing Code

```go
&schema.Schema{
    Type: schema.TypeMap,
    Elem: &schema.Schema{Type: schema.TypeString},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S006
&schema.Schema{
    Type: schema.TypeMap,
}
```
