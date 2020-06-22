# R013

The R013 analyzer reports cases of resource names which do not include at least one underscore character (`_`). Resources should be named with the provider name and API resource name separated by an underscore to clarify where a resource is declared and configured.

## Flagged Code

```go
map[string]*schema.Resource{
    "thing": /* ... */,
}
```

## Passing Code

```go
map[string]*schema.Resource{
    "example_thing": /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R013` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R013
map[string]*schema.Resource{
    "thing": /* ... */,
}
```
