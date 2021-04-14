# passes/helper/schema/resourceinfo

This pass only works with Terraform resources that are fully defined in a single function:

```go
func someResourceFunc() *schema.Resource {
    return &schema.Resource{ /* ... entire resource ... */ }
}
```
