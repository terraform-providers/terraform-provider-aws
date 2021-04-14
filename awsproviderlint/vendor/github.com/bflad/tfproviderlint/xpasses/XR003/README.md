# XR003

The XR003 analyzer reports missing usage of `Timeouts` in resources.

## Flagged Code

```go
&schema.Resource{
    Create: /* ... */,
    Delete: /* ... */,
    Read:   /* ... */,
    Schema: /* ... */,
}
```

## Passing Code

```go
&schema.Resource{
    Create:   /* ... */,
    Delete:   /* ... */,
    Read:     /* ... */,
    Schema:   /* ... */,
    Timeouts: &schema.ResourceTimeout{/* ... */},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR003` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR003
&schema.Resource{
    Create: /* ... */,
    Delete: /* ... */,
    Read: /* ... */,
}
```
