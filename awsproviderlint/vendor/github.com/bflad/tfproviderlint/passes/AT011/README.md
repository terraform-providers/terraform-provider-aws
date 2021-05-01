# AT011

The AT011 analyzer reports likely extraneous use of ID-only refresh testing. Most resources should prefer to include a `TestStep` with `ImportState` instead since it will cover the same testing functionality along with verifying resource import support.

However for cases where `IDRefreshName` is being already being used, the `IDRefreshIgnore` field is considered valid. If `IDRefreshName` is not being used, then this analyzer will return a report.

## Flagged Code

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
        IDRefreshIgnore: []string{"attr1"},
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
        PreCheck:      func() { testAccPreCheck(t) },
        Providers:     testAccProviders,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleThingConfig(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("example_thing.test", "attr1"),
                ),
            },
            {
                Config:            testAccExampleThingConfig(),
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

// ... or ...

func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:      func() { testAccPreCheck(t) },
        Providers:     testAccProviders,
        IDRefreshName: "example_thing.test",
        IDRefreshIgnore: []string{"attr1"},
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

Singular reports can be ignored by adding the a `//lintignore:AT011` Go code comment on the line immediately proceding, e.g.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
        //lintignore:AT011
        IDRefreshIgnore: []string{"attr1"},
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
