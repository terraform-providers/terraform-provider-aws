# AT004

The AT004 analyzer reports likely incorrect uses of `TestStep`
`Config` which define a provider configuration. Provider configurations should
be handled outside individual test configurations (e.g. environment variables).

## Flagged Code

```go
const ExampleThingConfig = `
provider "example" {}

resource "example_thing" {}
`

resource.TestCase{
    Steps: []resource.TestStep{
        {
            Config: ExampleThingConfig,
        },
    },
}
```

## Passing Code

```go
const ExampleThingConfig = `
resource "example_thing" {}
`

resource.TestCase{
    Steps: []resource.TestStep{
        {
            Config: ExampleThingConfig,
        },
    },
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT004` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT004
const ExampleThingConfig = `
provider "example" {}

resource "example_thing" {}
`
```
