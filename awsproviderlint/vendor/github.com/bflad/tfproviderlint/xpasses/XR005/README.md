# XR005

The XR005 analyzer reports cases of resources where `Description` is not configured, which is generally useful for providers that wish to automatically generate documentation based on the schema information.

This analyzer automatically ignores schema attribute `Elem` of type `schema.Resource`.

## Flagged Code

```go
&schema.Resource{
  Read:   /* ... */,
  Schema: map[string]*schema.Schema{/* ... */},
}
```

## Passing Code

```go
&schema.Resource{
  Description: "manages a widget",
  Read:        /* ... */,
  Schema:      map[string]*schema.Schema{/* ... */},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR005` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR005
&schema.Resource{
  Read:   /* ... */,
  Schema: map[string]*schema.Schema{/* ... */},
}
```
