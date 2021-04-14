# passes/helper/schema/schemamapcompositelit

This pass only works with Terraform schema maps that are fully defined:

```go
Schema: map[string]*schema.Schema{
    "attr": {
        Type:         schema.TypeString,
        Optional:     true,
        Computed:     true,
        ValidateFunc: /* ... */,
    },
},
```
