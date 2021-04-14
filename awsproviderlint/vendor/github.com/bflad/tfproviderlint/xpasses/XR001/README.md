# XR001

The XR001 analyzer reports usage of [`GetOkExists()`](https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ResourceData.GetOkExists) calls, which generally do not work as expected. Usage should be moved to standard `Get()` and `GetOk()` calls.

## Flagged Code

```go
d.GetOkExists("example")
```

## Passing Code

```go
d.Get("example")

// or

d.GetOk("example")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR001
d.GetOkExists("example")
```
