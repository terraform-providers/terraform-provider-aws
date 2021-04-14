# R018

The R018 analyzer reports [`time.Sleep()`](https://pkg.go.dev/time?tab=doc#Sleep) function usage. Terraform Providers should generally avoid this function when waiting for API operations and prefer polling methods such as [`resource.Retry()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/resource?tab=doc#Retry) or [`(resource.StateChangeConf).WaitForState()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/helper/resource?tab=doc#StateChangeConf.WaitForState). 

## Flagged Code

```go
time.Sleep(10)
```

## Passing Code

```go
err := resource.Retry(/* ... */)
```

Or

```go
stateConf := resource.StateChangeConf{/* ... */}
_, err := stateConf.WaitForState()
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R018` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R018
time.Sleep(10)
```
