# AT002

The AT002 analyzer reports where the string `_import` or `_Import` is used
in an acceptance test function name, which generally means there is an extraneous
acceptance test. `ImportState` testing should be included as a `TestStep` with each
applicable acceptance test, rather than a separate test that only verifies import
of a single test configuration.

## Flagged Code

```go
func TestAccExampleThing_basic(t *testing.T) { // not flagged!
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

func TestAccExampleThing_importBasic(t *testing.T) { // flagged!
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
            {
                ResourceName:      "example_thing.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

## Passing Code

```go
func TestAccExampleThing_basic(t *testing.T) {
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
            {
                ResourceName:      "example_thing.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT002
func TestAccExampleThing_importBasic(t *testing.T) {
    // ...
}
```
