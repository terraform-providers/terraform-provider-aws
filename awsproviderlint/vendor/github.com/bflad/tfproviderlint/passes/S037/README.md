# S037

The S037 analyzer reports cases of Schemas which include `ExactlyOneOf` and have invalid schema attribute references.

NOTE: This only verifies the syntax of attribute references. The Terraform Plugin SDK can unit test attribute references to verify the references against the full schema.

## Flagged Code

```go
&schema.Schema{
    ExactlyOneOf: []string{"config_block_attr.nested_attr"},
}
```

## Passing Code

```go
&schema.Schema{
    ExactlyOneOf: []string{"config_block_attr.0.nested_attr"},
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:S037` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:S037
&schema.Schema{
    ExactlyOneOf: []string{"config_block_attr.nested_attr"},
}
```
