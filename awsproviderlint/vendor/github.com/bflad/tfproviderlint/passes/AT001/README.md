# AT001

The AT001 analyzer reports likely incorrect uses of `TestCase`
which do not define a `CheckDestroy` function. `CheckDestroy` is used to verify
that test infrastructure has been removed at the end of an acceptance test.

Optional parameters:

`-ignored-filename-prefixes` Comma-separated list of filename prefixes to ignore, defaults to `data_source_`.  
`-ignored-filename-suffixes` Comma-separated list of filename suffixes to ignore, defaults to none.

More information can be found at:
https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy

## Flagged Code

NOTE: This analyzer does not differentiate between resource acceptance tests and data source acceptance tests. This is by design to ensure authors add the equivalent resource `CheckDestroy` function to data source testing, if available.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
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
func testAccCheckExampleThingDestroy(s *terraform.State) error {
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "example_thing" {
            continue
        }

        /* Code to check API for existence of Example Thing */
    }

    return nil
}

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

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT001` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
func TestAccExampleThing_Attr1(t *testing.T) {
    //lintignore:AT001
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
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
