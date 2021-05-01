# XR007

The XR007 analyzer reports usage of the [`os/exec.Command()`](https://pkg.go.dev/os/exec#Command) function. Providers that are using Go language based SDKs likely want to prevent any execution of other binaries for various reasons such as security and unexpected requirements (e.g. tool installation outside Terraform).

## Flagged Code

```go
var sneaky = exec.Command

sneaky("evilprogram")

exec.Command("evilprogram")
```

## Passing Code

```go
// Not present :)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR007` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR007
exec.Command("evilprogram")
```
