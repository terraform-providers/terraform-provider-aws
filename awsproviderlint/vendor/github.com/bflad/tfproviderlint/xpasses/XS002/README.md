# XS002

The XS002 analyzer reports cases of schemas where attributes are not listed in alphabetical order.

## Flagged Code

```go
map[string]*schema.Schema{
    "name": { /* ... */ },
    "arn": { /* ... */ },
  },
}
```

## Passing Code

```go
map[string]*schema.Schema{
    "arn": { /* ... */ },
    "name": { /* ... */ },
  },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XS002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XS002
map[string]*schema.Schema{
    "name": { /* ... */ },
    "arn": { /* ... */ },
  },
}
```
