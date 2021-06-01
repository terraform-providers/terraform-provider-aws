# R001

The R001 analyzer reports a complex key argument for a [`Set()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.Set)
call. It is preferred to explicitly use a string literal as the key argument.

## Flagged Code

```go
keys := []string{"example1", "example2"}
values := []string{"value1", "value2"}

for idx, key := range keys {
    d.Set(key, values[idx])
}
```

## Passing Code

```go
d.Set("example1", "value1")
d.Set("example2", "value2")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
keys := []string{"example1", "example2"}
values := []string{"value1", "value2"}

for idx, key := range keys {
    //lintignore:R001
    d.Set(key, values[idx])
}
```
