# S021

The S021 analyzer reports cases of schemas including `ComputedWhen`, which should be removed.

## Flagged Code

```go
&schema.Schema{
    Computed:     true,
    ComputedWhen: []string{"another_attr"},
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S021` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S021
&schema.Schema{
    Computed:     true,
    ComputedWhen: []string{"another_attr"},
}
```
