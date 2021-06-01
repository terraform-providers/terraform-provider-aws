# passes/helper/schema/schemainfo

This pass only works with Terraform schema that are fully defined:

```go
&schema.Schema{
    Type:         schema.TypeString,
    Optional:     true,
    Computed:     true,
    ValidateFunc: /* ... */,
},
```
