# XAT001

The XAT001 analyzer reports uses of `TestCase` which do not define an `ErrorCheck` function. `ErrorCheck` can be used to skip tests for known environmental issues.

## Flagged Code

NOTE: This analyzer does not differentiate between resource acceptance tests and data source acceptance tests. This is by design to ensure authors add the equivalent resource `CheckDestroy` function to data source testing, if available.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckExampleThingDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleThingConfig(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("example_thing.test", "attr1"),
                ),
            },
        },
    })
}
```

## Passing Code

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        ErrorCheck:   testAccErrorCheck,
        CheckDestroy: testAccCheckExampleThingDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleThingConfig(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("example_thing.test", "attr1"),
                ),
            },
        },
    })
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:XAT001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    //lintignore:XAT001
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckExampleThingDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleThingConfig(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("example_thing.test", "attr1"),
                ),
            },
        },
    })
}
```
