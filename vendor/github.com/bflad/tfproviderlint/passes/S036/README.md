# S036

The S036 analyzer reports cases of Schemas which include `ConflictsWith` and have invalid schema attribute references.

NOTE: This only verifies the syntax of attribute references. The Terraform Plugin SDK can unit test attribute references to verify the references against the full schema.

## Flagged Code

```go
&schema.Schema{
    ConflictsWith: []string{"config_block_attr.nested_attr"},
}
```

## Passing Code

```go
&schema.Schema{
    ConflictsWith: []string{"config_block_attr.0.nested_attr"},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S036` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S036
&schema.Schema{
    ConflictsWith: []string{"config_block_attr.nested_attr"},
}
```
