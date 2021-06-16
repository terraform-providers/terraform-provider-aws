# XR008

The XR008 analyzer reports usage of the [`os/exec.CommandContext()`](https://pkg.go.dev/os/exec#CommandContext) function. Providers that are using Go language based SDKs likely want to prevent any execution of other binaries for various reasons such as security and unexpected requirements (e.g. tool installation outside Terraform).

## Flagged Code

```go
var sneaky = exec.CommandContext

sneaky("evilprogram")

exec.CommandContext("evilprogram")
```

## Passing Code

```go
// Not present :)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XR008` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:XR008
exec.CommandContext("evilprogram")
```
