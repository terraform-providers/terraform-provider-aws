# S009

The S009 analyzer reports cases of `TypeList` or `TypeSet` schemas configuring `ValidateFunc`,
which will fail schema validation.

## Flagged Code

```go
&schema.Schema{
    Type:         schema.TypeList,
    Elem:         &schema.Schema{Type: schema.TypeString},
    ValidateFunc: /* ... */,
}

&schema.Schema{
    Type:         schema.TypeSet,
    Elem:         &schema.Schema{Type: schema.TypeString},
    ValidateFunc: /* ... */,
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

&schema.Schema{
    Type: schema.TypeList,
    Elem: &schema.Schema{
      Type:         schema.TypeString,
      ValidateFunc: /* ... */,
    },
}

&schema.Schema{
    Type: schema.TypeSet,
    Elem: &schema.Schema{
      Type:         schema.TypeString,
      ValidateFunc: /* ... */,
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S009` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S009
&schema.Schema{
    Type:         schema.TypeList,
    Elem:         &schema.Schema{Type: schema.TypeString},
    ValidateFunc: /* ... */,
}
```
