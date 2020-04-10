# XR002

The XR002 analyzer reports missing usage of `Importer` in resources.

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
    Importer: &schema.ResourceImporter{/* ... */},
    Read:     /* ... */,
    Schema:   /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR002
&schema.Resource{
    Create: /* ... */,
    Delete: /* ... */,
    Read: /* ... */,
}
```
