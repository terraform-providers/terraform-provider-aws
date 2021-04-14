# R011

The R011 analyzer reports cases of `Resource` which configure `MigrateState`. After Terraform 0.12, resources must configure new state migrations via `StateUpgraders`. Existing implementations of `MigrateState` prior to Terraform 0.12 can be ignored currently.

For additional information, see the [Extending Terraform documentation on state migrations](https://www.terraform.io/docs/extend/resources/state-migration.html).

## Flagged Code

```go
&schema.Resource{
    MigrateState:  /* ... */,
    SchemaVersion: 1,
}
```

## Passing Code

```go
&schema.Resource{
    SchemaVersion:  1,
    StateUpgraders: /* ... */,
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R011` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R011
&schema.Resource{
    MigrateState:  /* ... */,
    SchemaVersion: 1,
}
```
