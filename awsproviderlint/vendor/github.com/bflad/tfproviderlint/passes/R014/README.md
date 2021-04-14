# R014

The R014 analyzer reports when `CreateFunc`, `CreateContextFunc`, `DeleteFunc`, `DeleteContextFunc`, `ReadFunc`, `ReadContextFunc`, `UpdateFucn`, and `UpdateContextFunc` declarations do not use `d` as the name for the `*schema.ResourceData` parameter or `meta` as the name for the `interface{}` parameter. This parameter naming is the standard convention for resources.

## Flagged Code

```go
// Non-Context Functions

func resourceExampleThingCreate(invalid *schema.ResourceData, meta interface{}) error { /* ... */ }

func resourceExampleThingRead(d *schema.ResourceData, invalid interface{}) error { /* ... */ }

func resourceExampleThingDelete(invalid *schema.ResourceData, invalid interface{}) error { /* ... */ }

// Context Functions

func resourceExampleThingCreate(ctx context.Context, invalid *schema.ResourceData, meta interface{}) diag.Diagnostics { /* ... */ }

func resourceExampleThingRead(ctx context.Context, d *schema.ResourceData, invalid interface{}) diag.Diagnostics { /* ... */ }

func resourceExampleThingDelete(ctx context.Context, invalid *schema.ResourceData, invalid interface{}) diag.Diagnostics { /* ... */ }
```

## Passing Code

```go
// Non-Context Functions

func resourceExampleThingCreate(d *schema.ResourceData, meta interface{}) error { /* ... */ }

func resourceExampleThingRead(d *schema.ResourceData, meta interface{}) error { /* ... */ }

func resourceExampleThingDelete(d *schema.ResourceData, meta interface{}) error { /* ... */ }

// Context Functions

func resourceExampleThingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { /* ... */ }

func resourceExampleThingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { /* ... */ }

func resourceExampleThingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { /* ... */ }
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R014` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R014
func resourceExampleThingCreate(invalid *schema.ResourceData, meta interface{}) error { /* ... */ }
```
