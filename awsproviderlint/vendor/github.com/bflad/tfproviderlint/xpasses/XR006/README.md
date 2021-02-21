# XR006

The XR006 analyzer reports extraneous `Timeouts` fields in resources where the corresponding `Create`/`CreateContext`, `Delete`/`DeleteContext`, `Read`/`ReadContext`, or `Update`/`UpdateContext` implementation does not exist.

## Flagged Code

```go
&schema.Resource{
  /* ... no Create ... */
  Read: /* ... */,
  Timeouts: schema.ResourceTimesout{
    Create: schema.DefaultTimeout(10 * time.Minute),
  },
}
```

## Passing Code

```go
// Fixed Timeouts field alignment
&schema.Resource{
  /* ... no Create ... */
  Read: /* ... */,
  Timeouts: schema.ResourceTimesout{
    Read: schema.DefaultTimeout(10 * time.Minute),
  },
}

// Removed Timeouts
&schema.Resource{
  /* ... no Create ... */
  Read: /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR006
&schema.Resource{
  /* ... no Create ... */
  Read: /* ... */,
  Timeouts: schema.ResourceTimesout{
    Create: schema.DefaultTimeout(10 * time.Minute),
  },
}
```
