# S033

The S033 analyzer reports cases of schemas which enables only `Computed`
and configures `DefaultFunc`, which is not valid.

## Flagged Code

```go
&schema.Schema{
    Computed: true,
    DefaultFunc: /* ... */,
}
```

## Passing Code

```go
&schema.Schema{
    Computed: true,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S033` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S033
&schema.Schema{
    Computed: true,
    DefaultFunc: /* ... */,
}
```
