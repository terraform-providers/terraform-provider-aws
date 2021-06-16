# AT010

The AT010 analyzer reports likely extraneous use of ID-only refresh testing. Most resources should prefer to include a `TestStep` with `ImportState` instead since it will cover the same testing functionality along with verifying resource import support.

## Flagged Code

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:      func() { testAccPreCheck(t) },
        Providers:     testAccProviders,
        IDRefreshName: "example_thing.test",
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
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT010` Go code comment on the line immediately proceding, e.g.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:      func() { testAccPreCheck(t) },
        Providers:     testAccProviders,
        //lintignore:AT010
        IDRefreshName: "example_thing.test",
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
